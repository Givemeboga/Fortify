import requests
import secrets
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

    token = "fortify-" + secrets.token_hex(8)

    payloads = [p.replace("{TOKEN}", token) for p in PAYLOADS]

    params = list(inject_payload(url, payloads[0]).keys())

    for param in params:
        for payload in payloads:
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
        