import requests
import json
import re
from pathlib import Path

CONFIG_PATH = Path(__file__).parent.parent /"config" / "headers.json"

# Pulls "software/version" out of a header value, e.g. nginx/1.18.0, PHP/7.4.3
VERSION_RE = re.compile(r"([A-Za-z][\w.\-]*)/(\d+(?:\.\d+)+)")

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
        version_disclosures = {}

        for header in DEFENSIVE_HEADERS:
            if header in response.headers:
                present.append(header)
            else:
                missing.append(header)

        for header in LEAKY_HEADERS:
            if header in response.headers:
                value = response.headers[header]
                leaky[header] = response.headers[header]
                match = VERSION_RE.search(value)
                if match:
                    version_disclosures[header] = {
                        "header": header,
                        "software": match.group(1),
                        "version": match.group(2)
                    }

        return {
            "missing_headers": missing,
            "present_headers": present,
            "leaky_headers": leaky,
            "version_disclosures": version_disclosures,
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
            "version_disclosures": {},
            "redirect_info": {
                "final_url": None,
                "redirected": False,
                "chain": [],
            }
        }