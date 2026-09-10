import requests
from pathlib import Path

from scanner.active.injector import inject_payload

CONFIG_DIR = Path(__file__).parent.parent /"config"

PAYLOADS = []
with open(CONFIG_DIR / "xss_payloads.txt") as f:
    for line in f:
        cleaned = line.strip()
        if not cleaned:
            continue
        if cleaned.startswith("#"):
            continue
        PAYLOADS.append(cleaned)

def scan_xss(url: str) -> dict:
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
                if payload in response.text:
                    findings.append({
                        "parameter": param,
                        "payload": payload,
                        "url": injected_url
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
        