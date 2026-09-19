from fastapi import BackgroundTasks, FastAPI, HTTPException
from pydantic import BaseModel, HttpUrl

from db import init_db, create_scan, update_scan_results, get_scan, get_all_scans, update_scan_analysis, delete_scan, set_analysis_status

from scanner.passive.runner import run_passive_scan
from scanner.active.runner import run_active_scan

from analyzer.analyzer import analyze

from enum import Enum

class ScanType(str, Enum):
    passive = "passive"
    active = "active"

app = FastAPI()
init_db()
from fastapi.middleware.cors import CORSMiddleware

app.add_middleware(
    CORSMiddleware,
    allow_origins=["http://localhost:5173"],   # your dev frontend origin
    allow_methods=["*"],
    allow_headers=["*"],
)

class ScanRequest(BaseModel):
    url: HttpUrl
    scan_type: ScanType = ScanType.passive   # default to the safe option

def run_and_store(scan_id: str, url: str, scan_type: ScanType):
    try:
        if scan_type == ScanType.active:
            results = run_active_scan(url)
        else:
            results = run_passive_scan(url)
        update_scan_results(scan_id, results, "completed")
    except Exception as e:
        update_scan_results(scan_id, {"error": str(e)}, "failed")

def run_analysis(scan_id: str, provider: str | None, api_key: str | None):
    try:
        scan = get_scan(scan_id)                     # re-fetch to get results
        analysis = analyze(scan["results"], provider=provider, api_key=api_key)
        update_scan_analysis(scan_id, analysis)      # this sets status "completed"
    except Exception as e:
        # Background tasks fail invisibly otherwise — log the reason before marking failed.
        print(f"[run_analysis] analysis failed for {scan_id}: {e!r}")
        set_analysis_status(scan_id, "failed")

@app.post("/scan")
def start_scan(request: ScanRequest, background_tasks: BackgroundTasks):
    scan_id = create_scan(str(request.url), request.scan_type.value)

    background_tasks.add_task(run_and_store, scan_id, str(request.url), request.scan_type)

    return {"id": scan_id, "status": "pending"}

@app.get("/scans")
def list_scans():
    return get_all_scans()

class AnalyzeRequest(BaseModel):
    provider: str | None = None
    api_key: str | None = None

@app.post("/scans/{scan_id}/analyze")
def analyze_scan(scan_id: str, background_tasks: BackgroundTasks, body: AnalyzeRequest | None = None):
    scan = get_scan(scan_id)
    if scan is None:
        raise HTTPException(status_code=404, detail="Scan not found")
    if scan["status"] != "completed":
        raise HTTPException(status_code=409, detail="Scan not completed yet")

    provider = body.provider if body else None
    api_key = body.api_key if body else None

    set_analysis_status(scan_id, "analyzing")
    background_tasks.add_task(run_analysis, scan_id, provider, api_key)
    return {"id": scan_id, "analysis_status": "analyzing"}

@app.get("/scans/{scan_id}")
def read_scan(scan_id: str):
    scan = get_scan(scan_id)
    if scan is None:
        raise HTTPException(status_code=404, detail="Scan not found")
    return scan

@app.delete("/scans/{scan_id}")
def remove_scan(scan_id: str):
    scan = get_scan(scan_id)
    if scan is None:
        raise HTTPException(status_code=404, detail="Scan not found")
    delete_scan(scan_id)
    return {"deleted": scan_id}