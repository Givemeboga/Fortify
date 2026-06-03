import sqlite3
import json
from pathlib import Path
from datetime import datetime


DB_PATH = Path(__file__).parent / "fortify.db"

def get_connection():
    return sqlite3.connect(str(DB_PATH))

def init_db():
    conn = get_connection()
    cursor = conn.cursor()

    cursor.execute('''
        CREATE TABLE IF NOT EXISTS scans (
                   id INTEGER PRIMARY KEY AUTOINCREMENT,
                   target_url TEXT NOT NULL,
                   scan_type TEXT NOT NULL,
                   status TEXT NOT NULL,
                   results TEXT,
                   created_at TEXT,
                   completed_at TEXT
               )
    ''')

    conn.commit()
    conn.close()

def create_scan(target_url: str, scan_type: str) -> int:
    conn = get_connection()
    cursor = conn.cursor()

    now = datetime.utcnow().isoformat()

    cursor.execute(
        "INSERT INTO scans (target_url, scan_type, status, created_at) VALUES (?, ?, ?, ?)",
        (target_url, scan_type, "pending", now)
    )

    new_id = cursor.lastrowid

    conn.commit()
    conn.close() 

    return new_id

def update_scan_results(id: int, results: dict, status: str)  -> None:
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

def get_scan(id: int) -> dict | None:
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
        scans.append(scan)

    return scans