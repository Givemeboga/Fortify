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

BENIGN_FILE = Path(__file__).parent.parent / "config" / "benign_paths.txt"

BENIGN_PATHS = set()
with open(BENIGN_FILE) as f:
    for line in f:
        cleaned = line.strip()
        if not cleaned:
            continue
        if cleaned.startswith("#"):
            continue
        BENIGN_PATHS.add(cleaned)

# Fingerprints that a 200 HTML body is a real exposure (open directory listing
# or raw server-side source) rather than a normal application page. See issue #16.
LISTING_SIGNATURES = ["Index of /", "Parent Directory", "<?php", "-----BEGIN"]


def _probe(url: str):
    """GET a URL and return (status_code, body, content_type).
    status_code is None on failure."""
    try:
        # Using GET for reliable detection (some servers mishandle HEAD). See issue #2.
        response = requests.get(url, timeout=10)
        ctype = response.headers.get("content-type", "")
        return response.status_code, response.text, ctype
    except Exception:
        return None, "", ""


def _is_exposed(status, body, ctype) -> bool:
    """Decide whether a sensitive path is genuinely exposed, using content-level
    signals instead of status/length alone. Sites like GitHub return 200 for real
    pages (/administrator, /backup) that are NOT exposures — those are plain HTML
    with no smoking-gun signature, so we default them to safe. See issue #16."""
    if status != 200:
        return False
    if "text/html" not in ctype.lower():
        # A sensitive path serving non-HTML (text/plain, sql, octet-stream, …)
        # is almost always a real file leak, not an app page.
        return True
    if any(sig in body for sig in LISTING_SIGNATURES):
        # HTML, but with a directory-listing or raw-source fingerprint.
        return True
    # Plain HTML page, no signature — treat as a normal page, not an exposure.
    return False


def scan_paths(base: str) -> dict:
    base = base.rstrip("/")

    results = {}
    for path in SENSITIVE_PATHS:
        if path in BENIGN_PATHS:
            results[path] = {"status_code": 200, "exposed": False}
            continue
        status_code, body, content = _probe(base + path)
        results[path] = {
            "status_code": status_code, 
            "exposed": _is_exposed(status_code, body, content)
            }
    return results
