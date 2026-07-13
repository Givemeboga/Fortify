import requests
from pathlib import Path

from scanner.active.injector import inject_payload

CONFIG_DIR = Path(__file__).parent.parent /"config"

PAYLOADS = []
with open(CONFIG_DIR / "sqli_payloads.txt") as f:
    for line in f:
        cleaned = line.strip()
        if not cleaned:
            continue
        if cleaned.startswith("#"):
            continue
        PAYLOADS.append(cleaned)

SQL_ERRORS = []
with open(CONFIG_DIR / "sql_errors.txt") as f:
    for line in f:
        cleaned = line.strip()
        if not cleaned:
            continue
        if cleaned.startswith("#"):
            continue
        SQL_ERRORS.append(cleaned.lower())
        
def find_sql_error(body: str) -> str | None:
    for signature in SQL_ERRORS:
        if signature in body:
            return signature
    return None
        
def scan_sqli(url: str) ->dict:
    findings = []
    requests_made = 0
    errors = 0

    params = list(inject_payload(url, PAYLOADS[0]).keys())

    for param in params:
        for payload in PAYLOADS:
            injected_url = inject_payload(url, payload)[param]
            try:
                response = requests.get(injected_url, timeout=10)
                requests_made += 1
                matched = find_sql_error(response.text.lower())
                if matched:
                    findings.append({
                        "parameter": param,
                        "payload": payload,
                        "matched_error": matched,
                        })
                    break   # ← this param is confirmed; stop, move to next param
            except Exception:
                errors += 1
                continue

    return {
        "vulnerable": len(findings) > 0,
        "findings": findings,
        "requests_made": requests_made,
        "errors": errors,
        }