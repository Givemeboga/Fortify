package db

import (
	"database/sql"
	"time"
)

// Schedule is one Watchtower recurring scan: re-probe the target every
// interval_minutes, optionally auto-analyzing each run.
type Schedule struct {
	ID              string  `json:"id"`
	TargetURL       string  `json:"target_url"`
	ScanType        string  `json:"scan_type"`
	Ports           string  `json:"ports"`
	IntervalMinutes int     `json:"interval_minutes"`
	AutoAnalyze     bool    `json:"auto_analyze"`
	Provider        string  `json:"provider"`
	Model           string  `json:"model"`
	Enabled         bool    `json:"enabled"`
	CreatedAt       string  `json:"created_at"`
	LastRunAt       *string `json:"last_run_at"`
	NextRunAt       string  `json:"next_run_at"`
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// NextRun returns now+interval in DB time format (UTC RFC3339Nano, matching
// the lexicographic ORDERING the scans table already relies on).
func NextRun(from time.Time, intervalMinutes int) string {
	return from.Add(time.Duration(intervalMinutes) * time.Minute).UTC().Format(time.RFC3339Nano)
}

func CreateSchedule(s Schedule) (string, error) {
	mu.Lock()
	defer mu.Unlock()
	s.ID = newID()
	now := nowISO()
	next := NextRun(time.Now().UTC(), s.IntervalMinutes)
	_, err := conn.Exec(`INSERT INTO schedules
		(id, target_url, scan_type, ports, interval_minutes, auto_analyze, provider, model, enabled, created_at, next_run_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?)`,
		s.ID, s.TargetURL, s.ScanType, s.Ports, s.IntervalMinutes,
		boolInt(s.AutoAnalyze), s.Provider, s.Model, now, next,
	)
	if err != nil {
		return "", err
	}
	return s.ID, nil
}

func scheduleFromRow(id, targetURL, scanType, ports string, interval, autoAnalyze, enabled int,
	provider, model, created, nextRun string, lastRun sql.NullString) *Schedule {
	s := &Schedule{
		ID: id, TargetURL: targetURL, ScanType: scanType, Ports: ports,
		IntervalMinutes: interval, AutoAnalyze: autoAnalyze == 1, Enabled: enabled == 1,
		Provider: provider, Model: model, CreatedAt: created, NextRunAt: nextRun,
	}
	if lastRun.Valid && lastRun.String != "" {
		v := lastRun.String
		s.LastRunAt = &v
	}
	return s
}

func GetSchedule(id string) (*Schedule, error) {
	mu.Lock()
	defer mu.Unlock()
	var sid, url, stype, ports, provider, model, created, next string
	var interval, auto, enabled int
	var last sql.NullString
	err := conn.QueryRow(`SELECT id, target_url, scan_type, ports, interval_minutes, auto_analyze,
		enabled, provider, model, created_at, next_run_at, last_run_at
		FROM schedules WHERE id = ?`, id).
		Scan(&sid, &url, &stype, &ports, &interval, &auto, &enabled, &provider, &model, &created, &next, &last)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return scheduleFromRow(sid, url, stype, ports, interval, auto, enabled, provider, model, created, next, last), nil
}

func GetAllSchedules() ([]*Schedule, error) {
	mu.Lock()
	defer mu.Unlock()
	rows, err := conn.Query(`SELECT id, target_url, scan_type, ports, interval_minutes, auto_analyze,
		enabled, provider, model, created_at, next_run_at, last_run_at
		FROM schedules ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Schedule
	for rows.Next() {
		var sid, url, stype, ports, provider, model, created, next string
		var interval, auto, enabled int
		var last sql.NullString
		if err := rows.Scan(&sid, &url, &stype, &ports, &interval, &auto, &enabled,
			&provider, &model, &created, &next, &last); err != nil {
			return nil, err
		}
		out = append(out, scheduleFromRow(sid, url, stype, ports, interval, auto, enabled, provider, model, created, next, last))
	}
	if out == nil {
		out = []*Schedule{}
	}
	return out, rows.Err()
}

// GetDueSchedules returns enabled schedules whose next run has passed.
func GetDueSchedules(nowISO string) ([]*Schedule, error) {
	mu.Lock()
	defer mu.Unlock()
	rows, err := conn.Query(`SELECT id, target_url, scan_type, ports, interval_minutes, auto_analyze,
		enabled, provider, model, created_at, next_run_at, last_run_at
		FROM schedules WHERE enabled = 1 AND next_run_at <= ? ORDER BY next_run_at ASC`, nowISO)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Schedule
	for rows.Next() {
		var sid, url, stype, ports, provider, model, created, next string
		var interval, auto, enabled int
		var last sql.NullString
		if err := rows.Scan(&sid, &url, &stype, &ports, &interval, &auto, &enabled,
			&provider, &model, &created, &next, &last); err != nil {
			return nil, err
		}
		out = append(out, scheduleFromRow(sid, url, stype, ports, interval, auto, enabled, provider, model, created, next, last))
	}
	return out, rows.Err()
}

func MarkScheduleRun(id, lastRun, nextRun string) error {
	mu.Lock()
	defer mu.Unlock()
	_, err := conn.Exec(`UPDATE schedules SET last_run_at = ?, next_run_at = ? WHERE id = ?`,
		lastRun, nextRun, id)
	return err
}

// UpdateSchedule applies a partial update; a new interval re-anchors the
// next run to now so the change takes effect immediately.
func UpdateSchedule(id string, enabled *bool, intervalMinutes *int) (*Schedule, error) {
	mu.Lock()
	defer mu.Unlock()
	var curEnabled, curInterval int
	var last sql.NullString
	err := conn.QueryRow(`SELECT enabled, interval_minutes, last_run_at FROM schedules WHERE id = ?`, id).
		Scan(&curEnabled, &curInterval, &last)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if enabled != nil {
		curEnabled = boolInt(*enabled)
	}
	next := ""
	if intervalMinutes != nil {
		if *intervalMinutes < 5 {
			return nil, errInvalidInterval
		}
		curInterval = *intervalMinutes
		next = NextRun(time.Now().UTC(), curInterval)
	}
	if next != "" {
		_, err = conn.Exec(`UPDATE schedules SET enabled = ?, interval_minutes = ?, next_run_at = ? WHERE id = ?`,
			curEnabled, curInterval, next, id)
	} else {
		_, err = conn.Exec(`UPDATE schedules SET enabled = ? WHERE id = ?`, curEnabled, id)
	}
	if err != nil {
		return nil, err
	}
	mu.Unlock()
	sched, err := GetSchedule(id)
	mu.Lock()
	return sched, err
}

// errInvalidInterval is returned for intervals under the 5-minute floor.
var errInvalidInterval = errorString("interval_minutes must be >= 5")

type errorString string

func (e errorString) Error() string { return string(e) }

func DeleteSchedule(id string) error {
	mu.Lock()
	defer mu.Unlock()
	_, err := conn.Exec(`DELETE FROM schedules WHERE id = ?`, id)
	return err
}

// GetScansByTargetURL lists scans for one target, newest first — the basis
// for "previous scan" diffing.
func GetScansByTargetURL(targetURL string) ([]*Scan, error) {
	mu.Lock()
	defer mu.Unlock()
	rows, err := conn.Query(`SELECT id, target_url, scan_type, status, results, created_at,
		completed_at, analysis, analysis_status FROM scans WHERE target_url = ? ORDER BY created_at DESC`, targetURL)
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
