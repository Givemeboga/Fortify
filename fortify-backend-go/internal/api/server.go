// Package api implements the HTTP layer. Routes and JSON shapes match the
// Python FastAPI backend exactly so the React dashboard works unchanged:
//
//	POST /scan                 -> {id, status}
//	GET  /scans                -> [scan...]
//	GET  /scans/{id}           -> scan | 404
//	POST /scans/{id}/analyze   -> {id, analysis_status} | 404/409
//	DELETE /scans/{id}         -> {deleted} | 404
//	GET  /healthz              -> {status} (new, for Docker healthchecks)
package api

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"

	"fortify-go/internal/analyzer"
	"fortify-go/internal/db"
	"fortify-go/internal/scanner/active"
	"fortify-go/internal/scanner/passive"
	"fortify-go/internal/scanner/ports"
)

type ScanRequest struct {
	URL      string `json:"url"`
	ScanType string `json:"scan_type"`
	Ports    string `json:"ports"` // none (default) | top100 | top1000 | full
}

type AnalyzeRequest struct {
	Provider string `json:"provider"`
	APIKey   string `json:"api_key"`
	Model    string `json:"model"`
	BaseURL  string `json:"base_url"`
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeDetail(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"detail": msg})
}

// validateTarget mirrors FastAPI's HttpUrl + ScanType enum: http(s) URL with
// a host, scan_type passive|active (default passive). Failures are 422, same
// as the Python API boundary.
func validateTarget(raw, scanType string) (string, string, bool) {
	if scanType == "" {
		scanType = "passive"
	}
	if scanType != "passive" && scanType != "active" {
		return "", "", false
	}
	u, err := url.ParseRequestURI(strings.TrimSpace(raw))
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" {
		return "", "", false
	}
	return strings.TrimSpace(raw), scanType, true
}

func runAndStore(scanID, target, scanType, portProfile string) {
	var results map[string]any
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = errPanic(r)
			}
		}()
		type kv struct {
			k string
			v any
		}
		// Host-level port scan is independent of the web checks — run it
		// alongside instead of adding its seconds to the total.
		wantPorts := portProfile != "" && portProfile != ports.ProfileNone
		ch := make(chan kv, 2)
		go func() {
			if scanType == "active" {
				ch <- kv{"results", active.RunActiveScan(target)}
			} else {
				ch <- kv{"results", passive.RunPassiveScan(target)}
			}
		}()
		if wantPorts {
			go func() { ch <- kv{"ports", ports.ScanHost(target, portProfile)} }()
		}
		merged := map[string]any{}
		n := 1
		if wantPorts {
			n = 2
		}
		for range n {
			item := <-ch
			if item.k == "results" {
				for k, v := range item.v.(map[string]any) {
					merged[k] = v
				}
			} else {
				merged[item.k] = item.v
			}
		}
		results = merged
	}()
	if err != nil {
		_ = db.UpdateResults(scanID, map[string]any{"error": err.Error()}, "failed")
		return
	}
	if err := db.UpdateResults(scanID, results, "completed"); err != nil {
		log.Printf("[scan] update failed for %s: %v", scanID, err)
	}
}

func runAnalysis(scanID string, opts analyzer.LLMOptions) {
	scan, err := db.GetScan(scanID)
	if err != nil || scan == nil || scan.Results == nil {
		_ = db.FailAnalysis(scanID, "scan results unavailable")
		return
	}
	analysis, err := analyzer.Analyze(scan.Results, opts)
	if err != nil {
		log.Printf("[run_analysis] failed for %s: %v", scanID, err)
		_ = db.FailAnalysis(scanID, err.Error())
		return
	}
	// LLM-guardrail failure still persists as failed analysis (dashboard Retry).
	if _, hasErr := analysis["error"]; hasErr {
		raw, _ := json.Marshal(analysis)
		_ = db.FailAnalysis(scanID, string(raw))
		return
	}
	if err := db.UpdateAnalysis(scanID, analysis); err != nil {
		log.Printf("[run_analysis] store failed for %s: %v", scanID, err)
	}
}

func handleStartScan(w http.ResponseWriter, r *http.Request) {
	var req ScanRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		writeDetail(w, http.StatusUnprocessableEntity, "invalid request body")
		return
	}
	target, scanType, ok := validateTarget(req.URL, req.ScanType)
	if !ok {
		writeDetail(w, http.StatusUnprocessableEntity, "invalid url or scan_type (want http(s) url, passive|active)")
		return
	}
	portProfile := req.Ports
	if portProfile == "" {
		portProfile = ports.ProfileNone
	}
	if !ports.ValidProfile(portProfile) {
		writeDetail(w, http.StatusUnprocessableEntity, "invalid ports profile (want none|top100|top1000|full)")
		return
	}
	id, err := db.CreateScan(target, scanType)
	if err != nil {
		writeDetail(w, http.StatusInternalServerError, "could not create scan")
		return
	}
	go runAndStore(id, target, scanType, portProfile)
	writeJSON(w, http.StatusOK, map[string]string{"id": id, "status": "pending"})
}

func handleListScans(w http.ResponseWriter, r *http.Request) {
	scans, err := db.GetAllScans()
	if err != nil {
		writeDetail(w, http.StatusInternalServerError, "could not list scans")
		return
	}
	writeJSON(w, http.StatusOK, scans)
}

func handleReadScan(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	scan, err := db.GetScan(id)
	if err != nil {
		writeDetail(w, http.StatusInternalServerError, "could not read scan")
		return
	}
	if scan == nil {
		writeDetail(w, http.StatusNotFound, "Scan not found")
		return
	}
	writeJSON(w, http.StatusOK, scan)
}

func handleAnalyze(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	scan, err := db.GetScan(id)
	if err != nil {
		writeDetail(w, http.StatusInternalServerError, "could not read scan")
		return
	}
	if scan == nil {
		writeDetail(w, http.StatusNotFound, "Scan not found")
		return
	}
	if scan.Status != "completed" {
		writeDetail(w, http.StatusConflict, "Scan not completed yet")
		return
	}
	var body AnalyzeRequest
	_ = json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&body) // optional
	if err := db.SetAnalysisStatus(id, "analyzing"); err != nil {
		writeDetail(w, http.StatusInternalServerError, "could not start analysis")
		return
	}
	go runAnalysis(id, analyzer.LLMOptions{
		Provider: body.Provider,
		APIKey:   body.APIKey,
		Model:    body.Model,
		BaseURL:  body.BaseURL,
	})
	writeJSON(w, http.StatusOK, map[string]string{"id": id, "analysis_status": "analyzing"})
}

func handleDeleteScan(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	scan, err := db.GetScan(id)
	if err != nil {
		writeDetail(w, http.StatusInternalServerError, "could not read scan")
		return
	}
	if scan == nil {
		writeDetail(w, http.StatusNotFound, "Scan not found")
		return
	}
	if err := db.DeleteScan(id); err != nil {
		writeDetail(w, http.StatusInternalServerError, "could not delete scan")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"deleted": id})
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Same policy as the FastAPI CORSMiddleware: dev dashboard origin.
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// NewRouter wires the exact route table the dashboard expects.
func NewRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /scan", handleStartScan)
	mux.HandleFunc("GET /scans", handleListScans)
	mux.HandleFunc("GET /scans/{id}", handleReadScan)
	mux.HandleFunc("GET /scans/{id}/diff", handleScanDiff)
	mux.HandleFunc("POST /scans/{id}/analyze", handleAnalyze)
	mux.HandleFunc("DELETE /scans/{id}", handleDeleteScan)
	mux.HandleFunc("POST /schedules", handleCreateSchedule)
	mux.HandleFunc("GET /schedules", handleListSchedules)
	mux.HandleFunc("GET /schedules/{id}", handleReadSchedule)
	mux.HandleFunc("PATCH /schedules/{id}", handlePatchSchedule)
	mux.HandleFunc("DELETE /schedules/{id}", handleDeleteSchedule)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	return cors(mux)
}

type panicError struct{ v any }

func (e *panicError) Error() string { return "scan panicked" }

func errPanic(r any) error { return &panicError{v: r} }
