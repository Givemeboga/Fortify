import requests
import json
from pathlib import Path

CONFIG_PATH = Path(__file__).parent.parent /"config" / "headers.json"

with open(CONFIG_PATH) as f:
    config = json.load(f)

DEFENSIVE_HEADERS = config["defensive"]
LEAKY_HEADERS = config["leaky"]

def scan_headers(url: str) -> dict:
    try:
        response =  requests.get(url)

        redirected = bool(response.history)
        chain = [r.url for r in response.history]
        missing = []
        present = []
        leaky = {}

        for header in DEFENSIVE_HEADERS:
            if header in response.headers:
                present.append(header)
            else:
                missing.append(header)

        for header in LEAKY_HEADERS:
            if header in response.headers:
                leaky[header] = response.headers[header]

        return {
            "missing_headers": missing,
            "present_headers": present,
            "leaky_headers": leaky,
            "redirect_info": {
                "final_url": response.url,
                "redirected": redirected,
                "chain": chain
            }
        }

    except Exception as e:
        return {
            "missing_headers": [],
            "present_headers": [],
            "leaky_headers": {},
            "redirect_info": {
                "final_url": None,
                "redirected": False,
                "chain": [],
                }
        }