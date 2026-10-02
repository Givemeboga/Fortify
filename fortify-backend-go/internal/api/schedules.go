package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"fortify-go/internal/analyzer"
	"fortify-go/internal/db"
	"fortify-go/internal/diff"
)

type ScheduleRequest struct {
	URL             string `json:"url"`
	ScanType        string `json:"scan_type"`
	Ports           string `json:"ports"`
	IntervalMinutes int    `json:"interval_minutes"`
	AutoAnalyze     bool   `json:"auto_analyze"`
	Provider        string `json:"provider"`
	Model           string `json:"model"`
}

type SchedulePatch struct {
	Enabled         *bool `json:"enabled"`
	IntervalMinutes *int  `json:"interval_minutes"`
}

func handleCreateSchedule(w http.ResponseWriter, r *http.Request) {
	var req ScheduleRequest
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
		portProfile = "none"
	}
	if portProfile != "none" && portProfile != "top100" && portProfile != "top1000" && portProfile != "full" {
		writeDetail(w, http.StatusUnprocessableEntity, "invalid ports profile (want none|top100|top1000|full)")
		return
	}
	if req.IntervalMinutes < 5 {
		writeDetail(w, http.StatusUnprocessableEntity, "interval_minutes must be >= 5")
		return
	}
	if !analyzer.KnownProvider(req.Provider) {
		writeDetail(w, http.StatusUnprocessableEntity, "unknown provider (want ollama|gemini|openai|anthropic|custom)")
		return
	}
	id, err := db.CreateSchedule(db.Schedule{
		TargetURL: target, ScanType: scanType, Ports: portProfile,
		IntervalMinutes: req.IntervalMinutes, AutoAnalyze: req.AutoAnalyze,
		Provider: req.Provider, Model: req.Model,
	})
	if err != nil {
		writeDetail(w, http.StatusInternalServerError, "could not create schedule")
		return
	}
	sched, _ := db.GetSchedule(id)
	writeJSON(w, http.StatusCreated, sched)
}

func handleListSchedules(w http.ResponseWriter, r *http.Request) {
	scheds, err := db.GetAllSchedules()
	if err != nil {
		writeDetail(w, http.StatusInternalServerError, "could not list schedules")
		return
	}
	writeJSON(w, http.StatusOK, scheds)
}

func handleReadSchedule(w http.ResponseWriter, r *http.Request) {
	sched, err := db.GetSchedule(r.PathValue("id"))
	if err != nil {
		writeDetail(w, http.StatusInternalServerError, "could not read schedule")
		return
	}
	if sched == nil {
		writeDetail(w, http.StatusNotFound, "Schedule not found")
		return
	}
	writeJSON(w, http.StatusOK, sched)
}

func handlePatchSchedule(w http.ResponseWriter, r *http.Request) {
	var patch SchedulePatch
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&patch); err != nil {
		writeDetail(w, http.StatusUnprocessableEntity, "invalid request body")
		return
	}
	if patch.IntervalMinutes != nil && *patch.IntervalMinutes < 5 {
		writeDetail(w, http.StatusUnprocessableEntity, "interval_minutes must be >= 5")
		return
	}
	sched, err := db.UpdateSchedule(r.PathValue("id"), patch.Enabled, patch.IntervalMinutes)
	if err != nil {
		writeDetail(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	if sched == nil {
		writeDetail(w, http.StatusNotFound, "Schedule not found")
		return
	}
	writeJSON(w, http.StatusOK, sched)
}

func handleDeleteSchedule(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	sched, err := db.GetSchedule(id)
	if err != nil {
		writeDetail(w, http.StatusInternalServerError, "could not read schedule")
		return
	}
	if sched == nil {
		writeDetail(w, http.StatusNotFound, "Schedule not found")
		return
	}
	if err := db.DeleteSchedule(id); err != nil {
		writeDetail(w, http.StatusInternalServerError, "could not delete schedule")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"deleted": id})
}

// handleScanDiff compares a scan against another (default: the previous scan
// of the same target) and returns added/removed finding signatures.
func handleScanDiff(w http.ResponseWriter, r *http.Request) {
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
	againstID := r.URL.Query().Get("against")
	var against *db.Scan
	if againstID != "" {
		against, err = db.GetScan(againstID)
		if err != nil {
			writeDetail(w, http.StatusInternalServerError, "could not read comparison scan")
			return
		}
		if against == nil {
			writeDetail(w, http.StatusNotFound, "Comparison scan not found")
			return
		}
	} else {
		history, err := db.GetScansByTargetURL(scan.TargetURL)
		if err != nil {
			writeDetail(w, http.StatusInternalServerError, "could not list target scans")
			return
		}
		// Newest-first list: the first entry strictly older than this scan wins.
		for _, h := range history {
			if h.ID == scan.ID {
				continue
			}
			if h.CreatedAt < scan.CreatedAt || (h.CreatedAt == scan.CreatedAt && h.ID < scan.ID) {
				against = h
				break
			}
		}
		if against == nil {
			writeJSON(w, http.StatusOK, map[string]any{
				"id": id, "against": nil,
				"added": []string{}, "removed": []string{}, "note": "no previous scan for this target",
			})
			return
		}
	}
	oldRes := map[string]any{}
	if against.Results != nil {
		oldRes = against.Results
	}
	newRes := map[string]any{}
	if scan.Results != nil {
		newRes = scan.Results
	}
	d := diff.Diff(oldRes, newRes)
	writeJSON(w, http.StatusOK, map[string]any{
		"id": id, "against": against.ID,
		"added": d.Added, "removed": d.Removed,
		"added_count": len(d.Added), "removed_count": len(d.Removed),
	})
}

// --- Watchtower scheduler ---------------------------------------------------

// StartScheduler fires due schedules every 30s until stop is closed.
// Missed runs (server was down) simply become due and fire on the next tick.
func StartScheduler(stop <-chan struct{}) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	runDue() // catch up immediately on boot
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			runDue()
		}
	}
}

func runDue() {
	due, err := db.GetDueSchedules(time.Now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		log.Printf("[scheduler] due query failed: %v", err)
		return
	}
	for _, s := range due {
		scanID, err := db.CreateScan(s.TargetURL, s.ScanType)
		if err != nil {
			log.Printf("[scheduler] create scan failed for %s: %v", s.ID, err)
			continue
		}
		now := time.Now().UTC()
		if err := db.MarkScheduleRun(s.ID,
			now.Format(time.RFC3339Nano), db.NextRun(now, s.IntervalMinutes)); err != nil {
			log.Printf("[scheduler] mark run failed for %s: %v", s.ID, err)
		}
		log.Printf("[scheduler] firing %s → scan %s (%s)", s.ID, scanID, s.TargetURL)
		go runScheduled(s, scanID)
	}
}

// runScheduled executes the scan, then chains auto-analysis (when enabled)
// once the scan completes — polling with a bounded wait so a stuck scan
// can't pile up scheduler goroutines forever.
func runScheduled(s *db.Schedule, scanID string) {
	runAndStore(scanID, s.TargetURL, s.ScanType, s.Ports)
	if !s.AutoAnalyze {
		return
	}
	deadline := time.Now().Add(30 * time.Minute)
	for time.Now().Before(deadline) {
		time.Sleep(5 * time.Second)
		scan, err := db.GetScan(scanID)
		if err != nil || scan == nil {
			return
		}
		if scan.Status == "completed" {
			_ = db.SetAnalysisStatus(scanID, "analyzing")
			go runAnalysis(scanID, analyzer.LLMOptions{Provider: s.Provider, Model: s.Model})
			return
		}
		if scan.Status == "failed" {
			return
		}
	}
	log.Printf("[scheduler] scan %s did not complete in time; skipping auto-analysis", scanID)
}
