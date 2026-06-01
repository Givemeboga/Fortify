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
            # Using HEAD: no body = less bandwidth + stays passive.
            # Caveat: some servers mishandle HEAD; may need GET fallback later.
            response = requests.get(full_url)
            results[path] = {
                "status_code" : response.status_code,
            }
        except Exception as e:
            results[path] = {"status_code": None, "error": str(e)}
    return results