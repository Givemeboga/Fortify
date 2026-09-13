import json
from analyzer.llm import get_llm_response


def build_prompt(results: dict) -> str:
    return f"""You are a web application security analyst. 
You have been provided with the results of a security scan for a web application. Your task is to analyze the results and provide a detailed report highlighting any potential security vulnerabilities, their severity, and recommendations for remediation.
scan results:
{json.dumps(results, indent=2)}
Respond with ONLY a json object in exactly this structure:
{{
  "overall_risk": {{ "score": <0-100 int>, "level": "low|medium|high|critical" }},
  "summary": "<one paragraph>",
  "findings": [
    {{ "vulnerability": "...", "severity": {{"score": int, "level": "..."}},
      "explanation": "...", "remediation": "..." }}
  ],
  "priority_order": ["...", "..."]
}}"""

def analyze(results: dict) -> dict:
    prompt = build_prompt(results)
    raw = get_llm_response(prompt)
    try:
        return json.loads(raw)
    except json.JSONDecodeError:
        return {"error": "LLM returned invalid JSON", "raw": raw}