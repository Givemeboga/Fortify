import json
from analyzer.llm import get_llm_response

def build_prompt(findings: list[dict]) -> str:
    return f"""You are a senior application security analyst reviewing the output of an automated web scan.

You are given a list of CONFIRMED findings that the scanner has already verified. Treat them as ground truth: analyze these and ONLY these. Never invent, infer, generalize, merge, or add any finding, host, parameter, or vulnerability that is not explicitly in the list — even if you suspect one exists.

CONFIRMED FINDINGS (JSON):
{json.dumps(findings, indent=2)}

TASK — for every finding, produce:
- "vulnerability": a short, specific name for the issue.
- "explanation": 1-2 sentences on what the weakness is and the concrete risk it creates for THIS application; reference the exact parameter/header/path from the finding. No filler.
- "remediation": a specific, actionable fix — name the exact header and value, config directive, query change, or code pattern. Never write vague advice like "sanitize input" or "follow best practices".

THEN produce:
- "summary": 2-3 sentences of plain executive language naming the most important risk and its impact.
- "priority_order": the vulnerability names, most urgent first (most severe and most easily exploited first).

HARD RULES:
- Report ONLY the confirmed findings above. Do not add, merge, or split them.
- Do NOT assign severity scores or risk levels — those are computed separately in code. Focus only on explanation, remediation, summary, and ordering.
- Output a SINGLE json object and nothing else: no text before or after, no markdown code fences.

Respond in EXACTLY this structure:
{{
  "summary": "<2-3 sentences>",
  "findings": [
    {{ "vulnerability": "...", "explanation": "...", "remediation": "..." }}
  ],
  "priority_order": ["...", "..."]
}}"""

DISCLAIMER = "AI-generated guidance — verify findings before acting; the scan results are the authoritative facts."

SEVERITY_SCORES = {
    "sqli": 90,
    "sqli_boolean": 90,
    "path_traversal": 80,
    "exposed_path": 80,
    "xss": 75,
    "tls": 70,
    "missing_header": 50,
    "leaky_header": 15,
}

# Exposed paths that leak secrets/source — scored higher than a generic exposed path.
SECRET_PATHS = {"/.env", "/.git/", "/backup", "/db_backup.sql"}

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

    scores = [score_finding(f) for f in findings]   # deterministic, in finding order

    raw = get_llm_response(build_prompt(findings), provider=provider, api_key=api_key)
    try:
        assessment = json.loads(raw)
    except json.JSONDecodeError:
        return {"error": "LLM returned invalid JSON", "raw": raw}

    # Override the LLM's numbers with our code scores — the model provides prose only.
    for i, finding in enumerate(assessment.get("findings", [])):
        code_score = scores[i] if i < len(scores) else 30
        finding["severity"] = {"score": code_score, "level": score_to_level(code_score)}

    overall = max(scores)   # worst finding, not an average
    assessment["overall_risk"] = {"score": overall, "level": score_to_level(overall)}
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

def score_finding(finding: dict) -> int:
    score = SEVERITY_SCORES.get(finding["type"], 30)
    if finding["type"] == "leaky_header" and finding.get("version"):
        score += 10  # version leakage is more severe than generic leaky headers
    if finding["type"] == "exposed_path" and finding.get("path") in SECRET_PATHS:
        score += 10  # known sensitive paths are more severe than generic exposed paths
    return min(score, 100)  # cap at 100

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
        findings.append({
            "type": "leaky_header",
            "issue": f"Leaky header: {name} = {value}",
            "version": name in headers.get("version_disclosures", {}),
        })

    # --- Sensitive paths (only genuinely exposed ones — see issues #15/#16) ---
    # Public-by-design paths (robots.txt, …) are marked exposed=False by the
    # scanner itself now, so filtering here would be redundant.
    status = results.get("status", {})
    for path, info in status.items():
        if info.get("exposed"):
            findings.append({
                "type": "exposed_path",
                "issue": f"Accessible sensitive path: {path}",
                "path": path,
            })

    # --- Active checks ---
    for check in ("sqli", "sqli_boolean", "xss", "path_traversal"):
        section = results.get(check, {})
        if section.get("vulnerable"):
            for f in section.get("findings", []):
                findings.append({"type": check, "issue": f"{check} vulnerability in parameter '{f.get('parameter')}'"})

    return findings

