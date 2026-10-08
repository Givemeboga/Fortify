package analyzer

import (
	"encoding/json"
	"fmt"
	"strings"

	"fortify-go/internal/scanner/ports"
)

// strList tolerates DB-decoded ([]any of string) and in-process ([]string)
// string lists alike.
func strList(v any) []string {
	switch t := v.(type) {
	case []string:
		return t
	case []any:
		var out []string
		for _, item := range t {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}

// Disclaimer is stamped on every report: AI prose is guidance, scan facts win.
const Disclaimer = "AI-generated guidance — verify findings before acting; the scan results are the authoritative facts."

// SeverityScores: deterministic rule table. The LLM never scores — it only
// explains/remediates/summarizes. Parity with Python SEVERITY_SCORES.
var severityScores = map[string]int{
	"sqli":                90,
	"sqli_boolean":        90,
	"exposed_secret":      85,
	"path_traversal":      80,
	"exposed_path":        80,
	"xss":                 75,
	"tls":                 70,
	"cors_misconfig":      70,
	"risky_method":        55,
	"missing_header":      50,
	"cookie_no_secure":    50,
	"open_port":           40,
	"cookie_no_httponly":  40,
	"cookie_no_samesite":  30,
	"missing_security_txt": 20,
	"leaky_header":        15,
}

// SecretPaths leak credentials/source — scored higher than generic paths.
var secretPaths = map[string]bool{
	"/.env": true, "/.git/": true, "/backup": true, "/db_backup.sql": true,
}

func BuildPrompt(findings []map[string]any) string {
	raw, _ := json.MarshalIndent(findings, "", "  ")
	return fmt.Sprintf(`You are a senior application security analyst reviewing the output of an automated web scan.

You are given a list of CONFIRMED findings that the scanner has already verified. Treat them as ground truth: analyze these and ONLY these. Never invent, infer, generalize, merge, or add any finding, host, parameter, or vulnerability that is not explicitly in the list — even if you suspect one exists.

CONFIRMED FINDINGS (JSON):
%s

TASK — for every finding, produce:
- "id": echo the finding's "id" from the input EXACTLY (so scores stay attached to the right finding even if you reorder).
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
{
  "summary": "<2-3 sentences>",
  "findings": [
    { "id": "<echoed input id>", "vulnerability": "...", "explanation": "...", "remediation": "..." }
  ],
  "priority_order": ["...", "..."]
}`, string(raw))
}

// Analyze: extract findings in code -> score deterministically -> ask the LLM
// for prose only -> override with code scores. Empty findings skip the LLM
// entirely (nothing to invent) but still stamp provider/model.
func Analyze(results map[string]any, opts LLMOptions) (map[string]any, error) {
	findings := ExtractFindings(results)
	prov, model := ResolveProviderModel(opts.Provider, opts.Model)

	if len(findings) == 0 {
		return map[string]any{
			"overall_risk":   map[string]any{"score": 0, "level": "low"},
			"summary":        "No confirmed security findings were detected.",
			"findings":       []any{},
			"priority_order": []any{},
			"disclaimer":     Disclaimer,
			"provider":       prov,
			"model":          model,
		}, nil
	}

	scores := make([]int, len(findings))
	byID := make(map[string]int, len(findings))
	for i, f := range findings {
		scores[i] = ScoreFinding(f)
		// Stable ID echoed by the model, so scores attach by identity even
		// when the LLM reorders, merges, or drops findings.
		id := fmt.Sprintf("F%d", i+1)
		f["id"] = id
		byID[id] = scores[i]
	}

	raw, err := GetLLMResponse(BuildPrompt(findings), opts)
	if err != nil {
		return nil, err
	}
	var assessment map[string]any
	if err := extractJSON(raw, &assessment); err != nil {
		return map[string]any{"error": "LLM returned invalid JSON", "raw": raw}, nil
	}

	if list, ok := assessment["findings"].([]any); ok {
		applyScores(list, byID, scores)
	}
	overall := scores[0]
	for _, s := range scores[1:] {
		if s > overall {
			overall = s
		}
	}
	assessment["overall_risk"] = map[string]any{"score": overall, "level": ScoreToLevel(overall)}
	assessment["disclaimer"] = Disclaimer
	assessment["provider"] = prov
	assessment["model"] = model
	return assessment, nil
}

// applyScores stamps deterministic code scores onto the LLM's prose findings
// by echoed finding ID, falling back to input position only when the model
// drops the ID — a reordered or partial list can never shift scores around.
func applyScores(items []any, byID map[string]int, ordered []int) {
	for pos, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		score := 30
		if id, ok := m["id"].(string); ok {
			if s, found := byID[id]; found {
				score = s
			} else if pos < len(ordered) {
				score = ordered[pos]
			}
		} else if pos < len(ordered) {
			score = ordered[pos]
		}
		m["severity"] = map[string]any{"score": score, "level": ScoreToLevel(score)}
	}
}

// isRiskyPort tolerates every JSON number shape (int from in-process maps,
// float64 from DB-decoded JSON) when checking the risky-port table.
func isRiskyPort(v any) bool {
	switch n := v.(type) {
	case int:
		return ports.IsRisky(n)
	case int64:
		return ports.IsRisky(int(n))
	case float64:
		return ports.IsRisky(int(n))
	case json.Number:
		if i, err := n.Int64(); err == nil {
			return ports.IsRisky(int(i))
		}
	}
	return false
}

func portNumber(v any) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int64:
		return int(n), true
	case float64:
		return int(n), true
	case json.Number:
		if i, err := n.Int64(); err == nil {
			return int(i), true
		}
	}
	return 0, false
}

func ScoreToLevel(score int) string {
	switch {
	case score >= 90:
		return "critical"
	case score >= 70:
		return "high"
	case score >= 40:
		return "medium"
	default:
		return "low"
	}
}

func ScoreFinding(f map[string]any) int {
	t, _ := f["type"].(string)
	score, ok := severityScores[t]
	if !ok {
		score = 30
	}
	if t == "leaky_header" {
		if v, ok := f["version"].(bool); ok && v {
			score += 10
		}
	}
	if t == "exposed_path" {
		if p, ok := f["path"].(string); ok && secretPaths[p] {
			score += 10
		}
	}
	if t == "open_port" && isRiskyPort(f["port"]) {
		score = 75 // remote-admin / data-store exposure is high on its own
	}
	if score > 100 {
		score = 100
	}
	return score
}

// ExtractFindings pulls confirmed facts out of raw scan results in code — the
// grounding step that keeps the LLM honest. Parity with Python.
func ExtractFindings(results map[string]any) []map[string]any {
	var findings []map[string]any

	if tls, ok := results["tls"].(map[string]any); ok {
		if v, _ := tls["cert_expired"].(bool); v {
			findings = append(findings, map[string]any{"type": "tls", "issue": "TLS certificate is expired"})
		}
		if v, has := tls["cert_valid"]; has && v == false {
			findings = append(findings, map[string]any{"type": "tls", "issue": "TLS certificate is invalid"})
		}
	}

	if hdrs, ok := results["headers"].(map[string]any); ok {
		if missing, ok := hdrs["missing_headers"].([]any); ok {
			for _, h := range missing {
				if s, ok := h.(string); ok {
					findings = append(findings, map[string]any{"type": "missing_header", "issue": "Missing security header: " + s})
				}
			}
		}
		leaky, _ := hdrs["leaky_headers"].(map[string]any)
		disclosures, _ := hdrs["version_disclosures"].(map[string]any)
		if disclosures == nil {
			disclosures = map[string]any{}
		}
		for name, val := range leaky {
			_, hasVersion := disclosures[name]
			findings = append(findings, map[string]any{
				"type":    "leaky_header",
				"issue":   fmt.Sprintf("Leaky header: %s = %v", name, val),
				"version": hasVersion,
			})
		}
	}

	if status, ok := results["status"].(map[string]any); ok {
		for path, info := range status {
			if m, ok := info.(map[string]any); ok {
				if exposed, _ := m["exposed"].(bool); exposed {
					findings = append(findings, map[string]any{
						"type":  "exposed_path",
						"issue": "Accessible sensitive path: " + path,
						"path":  path,
					})
				}
			}
		}
	}

	for _, check := range []string{"sqli", "sqli_boolean", "xss", "path_traversal"} {
		if section, ok := results[check].(map[string]any); ok {
			if vuln, _ := section["vulnerable"].(bool); vuln {
				if list, ok := section["findings"].([]any); ok {
					for _, item := range list {
						param := ""
						if m, ok := item.(map[string]any); ok {
							param, _ = m["parameter"].(string)
						}
						findings = append(findings, map[string]any{
							"type":  check,
							"issue": fmt.Sprintf("%s vulnerability in parameter '%s'", check, param),
						})
					}
				}
			}
		}
	}

	// --- Open ports (only when a port profile was requested) ---
	if ps, ok := results["ports"].(map[string]any); ok {
		if hosts, ok := ps["hosts"].([]any); ok {
			for _, h := range hosts {
				hm, _ := h.(map[string]any)
				if hm == nil {
					continue
				}
				ip, _ := hm["ip"].(string)
				if list, ok := hm["open_ports"].([]any); ok {
					for _, item := range list {
						om, _ := item.(map[string]any)
						if om == nil {
							continue
						}
						port, ok := portNumber(om["port"])
						if !ok {
							continue
						}
						svc, _ := om["service"].(string)
						if svc == "" {
							svc = "unknown"
						}
						where := fmt.Sprintf("%d/tcp (%s)", port, svc)
						if ip != "" {
							where += " on " + ip
						}
						findings = append(findings, map[string]any{
							"type":  "open_port",
							"issue": "Open port: " + where,
							"port":  port,
						})
					}
				}
			}
		}
	}

	// --- Cookie flags (one finding per missing-flag category) ---
	if cs, ok := results["cookies"].(map[string]any); ok {
		for _, cat := range []struct {
			key, ftype, label string
		}{
			{"missing_secure", "cookie_no_secure", "missing Secure flag"},
			{"missing_httponly", "cookie_no_httponly", "missing HttpOnly flag"},
			{"missing_samesite", "cookie_no_samesite", "missing SameSite policy"},
		} {
			if names := strList(cs[cat.key]); len(names) > 0 {
				findings = append(findings, map[string]any{
					"type":  cat.ftype,
					"issue": fmt.Sprintf("%d cookie(s) %s: %s", len(names), cat.label, strings.Join(names, ", ")),
				})
			}
		}
	}

	// --- CORS ---
	if co, ok := results["cors"].(map[string]any); ok {
		if bad, _ := co["misconfigured"].(bool); bad {
			detail, _ := co["detail"].(string)
			if detail == "" {
				detail = "Cross-origin trust allows an attacker origin with credentials"
			}
			findings = append(findings, map[string]any{"type": "cors_misconfig", "issue": detail})
		}
	}

	// --- HTTP methods ---
	if ms, ok := results["methods"].(map[string]any); ok {
		if risky := strList(ms["risky"]); len(risky) > 0 {
			findings = append(findings, map[string]any{
				"type":  "risky_method",
				"issue": "Risky HTTP methods enabled: " + strings.Join(risky, ", "),
			})
		}
	}

	// --- security.txt ---
	if st, ok := results["security_txt"].(map[string]any); ok {
		if present, _ := st["present"].(bool); !present {
			findings = append(findings, map[string]any{
				"type":  "missing_security_txt",
				"issue": "No vulnerability-disclosure channel (/.well-known/security.txt missing or without Contact)",
			})
		}
	}

	// --- Client-side secrets ---
	if ss, ok := results["secrets"].(map[string]any); ok {
		if list, ok := ss["findings"].([]any); ok {
			for _, item := range list {
				m, _ := item.(map[string]any)
				if m == nil {
					continue
				}
				kind, _ := m["kind"].(string)
				loc, _ := m["location"].(string)
				findings = append(findings, map[string]any{
					"type":  "exposed_secret",
					"issue": fmt.Sprintf("Exposed secret (%s) in %s", kind, loc),
					"kind":  kind,
				})
			}
		}
	}

	if findings == nil {
		findings = []map[string]any{}
	}
	return findings
}
