import json
from analyzer.llm import get_llm_response

def build_prompt(findings: list[dict]) -> str:
    return f"""You are a web application security analyst.

Below is a list of CONFIRMED security findings from an automated scan. These are the ONLY
issues you may report — do NOT invent, infer, or add any finding not in this list.

Confirmed findings:
{json.dumps(findings, indent=2)}

For each confirmed finding, write a clear explanation, assign a numeric severity score
(0-100), and give a concrete remediation. Then produce an overall risk score and a
priority order (most urgent first).

Rules:
- Report ONLY the findings listed above. Do not add new ones.
- Provide only numeric scores (0-100). Do NOT include risk levels — computed separately.

Respond with ONLY a json object in exactly this structure:
{{
  "overall_risk": {{ "score": <0-100 int> }},
  "summary": "<one paragraph>",
  "findings": [
    {{ "vulnerability": "...", "severity": {{ "score": <0-100 int> }},
      "explanation": "...", "remediation": "..." }}
  ],
  "priority_order": ["...", "..."]
}}"""

DISCLAIMER = "AI-generated guidance — verify findings before acting; the scan results are the authoritative facts."

def analyze(results: dict, provider: str | None = None, api_key: str | None = None) -> dict:
    findings = extract_findings(results)

    # No confirmed findings → don't even call the LLM (nothing to invent)
    if not findings:
        return {
            "overall_risk": {"score": 0, "level": "low"},
            "summary": "No confirmed security findings were detected.",
            "findings": [],
            "priority_order": [],
            "disclaimer": DISCLAIMER,
        }

    raw = get_llm_response(build_prompt(findings), provider=provider, api_key=api_key)
    try:
        assessment = json.loads(raw)
    except json.JSONDecodeError:
        return {"error": "LLM returned invalid JSON", "raw": raw}

    assessment["overall_risk"]["level"] = score_to_level(assessment["overall_risk"]["score"])
    for finding in assessment.get("findings", []):
        finding["severity"]["level"] = score_to_level(finding["severity"]["score"])
    assessment["disclaimer"] = DISCLAIMER
    return assessment

def score_to_level(score: int) -> str:
    if score >= 90:
        return "critical"
    elif score >= 70:
        return "high"
    elif score >= 40:
        return "medium"
    else:
        return "low"

def extract_findings(results: dict) -> list[dict]:
    findings = []

    # --- TLS ---
    tls = results.get("tls", {})
    if tls.get("cert_expired"):
        findings.append({"type": "tls", "issue": "TLS certificate is expired"})
    if tls.get("cert_valid") is False:
        findings.append({"type": "tls", "issue": "TLS certificate is invalid"})
    # cipher evaluation intentionally postponed (needs a known-weak-cipher list)

    # --- Headers ---
    headers = results.get("headers", {})
    for h in headers.get("missing_headers", []):
        findings.append({"type": "missing_header", "issue": f"Missing security header: {h}"})
    for name, value in headers.get("leaky_headers", {}).items():
        findings.append({"type": "leaky_header", "issue": f"Leaky header: {name} = {value}"})

    # --- Sensitive paths (only genuinely exposed ones — see issues #15/#16) ---
    # Public-by-design paths (robots.txt, …) are marked exposed=False by the
    # scanner itself now, so filtering here would be redundant.
    status = results.get("status", {})
    for path, info in status.items():
        if info.get("exposed"):
            findings.append({"type": "exposed_path", "issue": f"Accessible sensitive path: {path}"})

    # --- Active checks ---
    for check in ("sqli", "xss", "path_traversal"):
        section = results.get(check, {})
        if section.get("vulnerable"):
            for f in section.get("findings", []):
                findings.append({"type": check, "issue": f"{check} vulnerability in parameter '{f.get('parameter')}'"})

    return findings

