import requests
from pathlib import Path

PATHS_FILE = Path(__file__).parent.parent / "config" / "paths.txt"

SENSITIVE_PATHS = []
with open(PATHS_FILE) as f:
    for line in f:
        cleaned = line.strip()
        if not cleaned:
            continue
        if cleaned.startswith("#"):
            continue
        SENSITIVE_PATHS.append(cleaned)

def scan_paths(base: str) -> dict:
    results = {}

    for path in SENSITIVE_PATHS:
        full_url = base.rstrip("/") + path

        try:
            # Using GET for reliable detection (some servers mishandle HEAD).
            # Body is intentionally discarded to stay passive. See issue #2 for HEAD optimization.
            response = requests.get(full_url)
            results[path] = {
                "status_code" : response.status_code,
            }
        except Exception as e:
            results[path] = {"status_code": None, "error": str(e)}
    return results