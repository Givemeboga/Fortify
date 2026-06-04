from fastapi import BackgroundTasks, FastAPI, HTTPException
from pydantic import BaseModel, HttpUrl

from db import init_db, create_scan, update_scan_results, get_scan, get_all_scans
from scanner.passive.runner import run_passive_scan

app = FastAPI()
init_db()

class ScanRequest(BaseModel):
    url: HttpUrl

def run_and_store(scan_id: int, url: str):
    try:
        results = run_passive_scan(url)
        update_scan_results(scan_id, results, "completed")
    except Exception as e:
        update_scan_results(scan_id, {"error": str(e)}, "failed")

@app.post("/scan")
def start_scan(request: ScanRequest, background_tasks: BackgroundTasks):
    scan_id = create_scan(str(request.url), "passive")

    background_tasks.add_task(run_and_store, scan_id, str(request.url))

    return {"id": scan_id, "status": "pending"}

@app.get("/scans")
def list_scans():
    return get_all_scans()

@app.get("/scans/{scan_id}")
def read_scan(scan_id: int):
    scan = get_scan(scan_id)

    if scan is None:
        raise HTTPException(status_code=404, detail="Scan not found")
    return scan