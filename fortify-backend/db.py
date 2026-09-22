import sqlite3
import json
from pathlib import Path
from datetime import datetime
import secrets
import os


DB_PATH = Path(os.getenv("FORTIFY_DB_PATH", str(Path(__file__).parent / "fortify.db")))

def get_connection():
    return sqlite3.connect(str(DB_PATH))

def init_db():
    conn = get_connection()
    cursor = conn.cursor()

    cursor.execute('''
        CREATE TABLE IF NOT EXISTS scans (
                   id TEXT PRIMARY KEY,
                   target_url TEXT NOT NULL,
                   scan_type TEXT NOT NULL,
                   status TEXT NOT NULL,
                   results TEXT,
                   created_at TEXT,
                   completed_at TEXT,
                   analysis TEXT,
                   analysis_status TEXT
               )
    ''')

    conn.commit()
    conn.close()

def create_scan(target_url: str, scan_type: str) -> str:
    conn = get_connection()
    cursor = conn.cursor()

    now = datetime.utcnow().isoformat()

    new_id = secrets.token_urlsafe(16)  # Generate a unique ID for the scan
    cursor.execute(
        "INSERT INTO scans (id, target_url, scan_type, status, created_at) VALUES (?, ?, ?, ?, ?)",
        (new_id, target_url, scan_type, "pending", now)
    )

    conn.commit()
    conn.close() 

    return new_id

def update_scan_results(id: str, results: dict, status: str)  -> None:
    conn = get_connection()
    cursor = conn.cursor()

    completed = datetime.utcnow().isoformat()

    results_json = json.dumps(results)

    cursor.execute(
        "UPDATE scans SET results = ?, status = ?, completed_at = ? WHERE id = ?",
        (results_json, status, completed, id)
    )

    conn.commit()
    conn.close()

def get_scan(id: str) -> dict | None:
    conn = get_connection()
    conn.row_factory = sqlite3.Row
    cursor = conn.cursor()

    cursor.execute("SELECT * FROM scans WHERE id = ?", (id,))
    row = cursor.fetchone()

    conn.close()

    if row is None:
        return None
    
    scan = dict(row)
    
    if scan["results"] is not None:
        scan["results"] = json.loads(scan["results"])
    if scan["analysis"] is not None:
         scan["analysis"] = json.loads(scan["analysis"])
    
    return scan

def get_all_scans() -> list[dict]:
    conn = get_connection()
    conn.row_factory = sqlite3.Row
    cursor = conn.cursor()

    cursor.execute("SELECT * FROM scans ORDER BY created_at DESC")
    rows = cursor.fetchall()

    scans = []
    for row in rows:
        scan = dict(row)
        if scan["results"] is not None:
            scan["results"] = json.loads(scan["results"])
        if scan["analysis"] is not None:
            scan["analysis"] = json.loads(scan["analysis"])
        scans.append(scan)

    return scans

def update_scan_analysis(id: str, analysis: dict) -> None:
    conn = get_connection()
    cursor = conn.cursor()
    cursor.execute(
        "UPDATE scans SET analysis = ?, analysis_status = ? WHERE id = ?",
        (json.dumps(analysis), "completed", id)   # serialize dict → JSON string, same as results
    )
    conn.commit()
    conn.close()

def delete_scan(id: str) -> None:
    conn = get_connection()
    cursor = conn.cursor()
    cursor.execute("DELETE FROM scans WHERE id = ?", (id,))
    conn.commit()
    conn.close()

def set_analysis_status(id: str, status: str) -> None:
    conn = get_connection()
    cursor = conn.cursor()
    cursor.execute(
        "UPDATE scans SET analysis_status = ? WHERE id = ?",
        (status, id)
    )
    conn.commit()
    conn.close()

def fail_analysis(id: str, error_message: str) -> None:
    conn = get_connection()
    cursor = conn.cursor()
    cursor.execute(
        "UPDATE scans SET analysis_status = ?, analysis = ? WHERE id = ?",
        ("failed", json.dumps({"error": error_message}), id)
    )
    conn.commit()
    conn.close()