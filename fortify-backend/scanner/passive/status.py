import requests
import secrets
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


def _probe(url: str):
    """GET a URL and return (status_code, body_length). Status is None on failure."""
    try:
        # Using GET for reliable detection (some servers mishandle HEAD). See issue #2.
        response = requests.get(url, timeout=10)
        return response.status_code, len(response.text)
    except Exception:
        return None, 0


def _is_exposed(status, length, base_status, base_len) -> bool:
    """Decide whether a path is genuinely exposed, using a baseline probe to
    filter out sites that return 200 for arbitrary paths (SPAs, catch-alls,
    user-profile routes). See issue #15."""
    if status != 200:
        return False
    if base_status == 200:
        # The site returns 200 even for paths that don't exist, so a 200 alone is
        # meaningless — only flag when the content differs clearly from that baseline.
        return abs(length - base_len) > max(512, base_len * 0.5)
    # A normal server 404s missing paths, so a real 200 here is a genuine hit.
    return True


def scan_paths(base: str) -> dict:
    base = base.rstrip("/")

    # Baseline: a path that should NOT exist. Tells us how the site answers "not found".
    baseline_url = base + "/fortify-nonexistent-" + secrets.token_hex(8)
    base_status, base_len = _probe(baseline_url)

    results = {}
    for path in SENSITIVE_PATHS:
        status, length = _probe(base + path)
        results[path] = {
            "status_code": status,
            "exposed": _is_exposed(status, length, base_status, base_len),
        }
    return results
