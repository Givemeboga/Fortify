package db

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"crypto/rand"
	"encoding/base64"

	_ "modernc.org/sqlite"
)

var (
	conn *sql.DB
	mu   sync.Mutex
)

// Scan mirrors the row shape the React dashboard consumes.
type Scan struct {
	ID             string         `json:"id"`
	TargetURL      string         `json:"target_url"`
	ScanType       string         `json:"scan_type"`
	Status         string         `json:"status"`
	Results        map[string]any `json:"results"`
	CreatedAt      string         `json:"created_at"`
	CompletedAt    *string        `json:"completed_at"`
	Analysis       map[string]any `json:"analysis"`
	AnalysisStatus *string        `json:"analysis_status"`
}

func dbPath() string {
	if p := os.Getenv("FORTIFY_DB_PATH"); p != "" {
		return p
	}
	exe, err := os.Executable()
	if err == nil {
		return filepath.Join(filepath.Dir(exe), "fortify.db")
	}
	return "fortify.db"
}

// Init opens the SQLite DB (WAL mode for concurrent readers + one writer)
// and creates the scans table. Parity with Python db.py.
func Init() error {
	mu.Lock()
	defer mu.Unlock()
	if conn != nil {
		return nil
	}
	path := dbPath()
	if dir := filepath.Dir(path); dir != "." && dir != "" {
		_ = os.MkdirAll(dir, 0o755)
	}
	c, err := sql.Open("sqlite", path)
	if err != nil {
		return err
	}
	c.SetMaxOpenConns(1) // SQLite single-writer; readers serialize through it
	for _, pragma := range []string{
		`PRAGMA journal_mode = WAL`,
		`PRAGMA busy_timeout = 5000`,
	} {
		if _, err := c.Exec(pragma); err != nil {
			return err
		}
	}
	if _, err := c.Exec(`CREATE TABLE IF NOT EXISTS scans (
		id TEXT PRIMARY KEY,
		target_url TEXT NOT NULL,
		scan_type TEXT NOT NULL,
		status TEXT NOT NULL,
		results TEXT,
		created_at TEXT,
		completed_at TEXT,
		analysis TEXT,
		analysis_status TEXT
	)`); err != nil {
		return err
	}
	// Watchtower: recurring scans. Booleans ride as INTEGER 0/1.
	if _, err := c.Exec(`CREATE TABLE IF NOT EXISTS schedules (
		id TEXT PRIMARY KEY,
		target_url TEXT NOT NULL,
		scan_type TEXT NOT NULL,
		ports TEXT NOT NULL DEFAULT 'none',
		interval_minutes INTEGER NOT NULL,
		auto_analyze INTEGER NOT NULL DEFAULT 0,
		provider TEXT NOT NULL DEFAULT '',
		model TEXT NOT NULL DEFAULT '',
		enabled INTEGER NOT NULL DEFAULT 1,
		created_at TEXT,
		last_run_at TEXT,
		next_run_at TEXT
	)`); err != nil {
		return err
	}
	conn = c
	return nil
}

func newID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		// fallback: timestamp-based (practically unreachable)
		return "scan-" + time.Now().UTC().Format("20060102150405.000000000")
	}
	return base64.RawURLEncoding.EncodeToString(b[:]) // 22 chars, unguessable like token_urlsafe(16)
}

func nowISO() string { return time.Now().UTC().Format(time.RFC3339Nano) }

func CreateScan(targetURL, scanType string) (string, error) {
	mu.Lock()
	defer mu.Unlock()
	id := newID()
	_, err := conn.Exec(
		`INSERT INTO scans (id, target_url, scan_type, status, created_at) VALUES (?, ?, ?, ?, ?)`,
		id, targetURL, scanType, "pending", nowISO(),
	)
	if err != nil {
		return "", err
	}
	return id, nil
}

func UpdateResults(id string, results map[string]any, status string) error {
	mu.Lock()
	defer mu.Unlock()
	raw, err := json.Marshal(results)
	if err != nil {
		return err
	}
	done := nowISO()
	_, err = conn.Exec(`UPDATE scans SET results = ?, status = ?, completed_at = ? WHERE id = ?`,
		string(raw), status, done, id)
	return err
}

func scanFromRow(id, targetURL, scanType, status string, results sql.NullString, createdAt string, completedAt, analysis, analysisStatus sql.NullString) (*Scan, error) {
	s := &Scan{
		ID: id, TargetURL: targetURL, ScanType: scanType,
		Status: status, CreatedAt: createdAt,
	}
	if results.Valid && results.String != "" {
		var m map[string]any
		if err := json.Unmarshal([]byte(results.String), &m); err == nil {
			s.Results = m
		}
	}
	if completedAt.Valid && completedAt.String != "" {
		v := completedAt.String
		s.CompletedAt = &v
	}
	if analysis.Valid && analysis.String != "" {
		var m map[string]any
		if err := json.Unmarshal([]byte(analysis.String), &m); err == nil {
			s.Analysis = m
		}
	}
	if analysisStatus.Valid && analysisStatus.String != "" {
		v := analysisStatus.String
		s.AnalysisStatus = &v
	}
	return s, nil
}

func GetScan(id string) (*Scan, error) {
	mu.Lock()
	defer mu.Unlock()
	var sid, url, stype, status, created string
	var results, completed, analysis, astatus sql.NullString
	err := conn.QueryRow(`SELECT id, target_url, scan_type, status, results, created_at, completed_at, analysis, analysis_status FROM scans WHERE id = ?`, id).
		Scan(&sid, &url, &stype, &status, &results, &created, &completed, &analysis, &astatus)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return scanFromRow(sid, url, stype, status, results, created, completed, analysis, astatus)
}

func GetAllScans() ([]*Scan, error) {
	mu.Lock()
	defer mu.Unlock()
	rows, err := conn.Query(`SELECT id, target_url, scan_type, status, results, created_at, completed_at, analysis, analysis_status FROM scans ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Scan
	for rows.Next() {
		var sid, url, stype, status, created string
		var results, completed, analysis, astatus sql.NullString
		if err := rows.Scan(&sid, &url, &stype, &status, &results, &created, &completed, &analysis, &astatus); err != nil {
			return nil, err
		}
		s, err := scanFromRow(sid, url, stype, status, results, created, completed, analysis, astatus)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	if out == nil {
		out = []*Scan{}
	}
	return out, rows.Err()
}

func UpdateAnalysis(id string, analysis map[string]any) error {
	mu.Lock()
	defer mu.Unlock()
	raw, err := json.Marshal(analysis)
	if err != nil {
		return err
	}
	_, err = conn.Exec(`UPDATE scans SET analysis = ?, analysis_status = ? WHERE id = ?`,
		string(raw), "completed", id)
	return err
}

func DeleteScan(id string) error {
	mu.Lock()
	defer mu.Unlock()
	_, err := conn.Exec(`DELETE FROM scans WHERE id = ?`, id)
	return err
}

func SetAnalysisStatus(id, status string) error {
	mu.Lock()
	defer mu.Unlock()
	_, err := conn.Exec(`UPDATE scans SET analysis_status = ? WHERE id = ?`, status, id)
	return err
}

func FailAnalysis(id, msg string) error {
	mu.Lock()
	defer mu.Unlock()
	raw, _ := json.Marshal(map[string]any{"error": msg})
	_, err := conn.Exec(`UPDATE scans SET analysis_status = ?, analysis = ? WHERE id = ?`, "failed", string(raw), id)
	return err
}
