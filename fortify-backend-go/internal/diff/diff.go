// Package diff computes finding-level changes between two scans of the same
// target ("port 3389 opened since yesterday"). Signatures are derived from
// raw results with the same rules the analyzer uses, so the diff speaks the
// same language as the Battle Report.
package diff

import (
	"fmt"
	"sort"
)

// strList tolerates DB-decoded ([]any) and in-process ([]string) lists.
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

func portStr(v any) string {
	switch n := v.(type) {
	case int:
		return fmt.Sprintf("%d", n)
	case int64:
		return fmt.Sprintf("%d", n)
	case float64:
		return fmt.Sprintf("%d", int(n))
	}
	return ""
}

// Signatures reduces raw scan results to a sorted set of stable finding keys.
func Signatures(results map[string]any) []string {
	set := map[string]bool{}
	add := func(s string) { set[s] = true }

	if tls, ok := results["tls"].(map[string]any); ok {
		if v, _ := tls["cert_expired"].(bool); v {
			add("tls:expired")
		}
		if v, has := tls["cert_valid"]; has && v == false {
			add("tls:invalid")
		}
	}
	if hdrs, ok := results["headers"].(map[string]any); ok {
		for _, h := range strList(hdrs["missing_headers"]) {
			add("missing_header:" + h)
		}
		if leaky, ok := hdrs["leaky_headers"].(map[string]any); ok {
			for name := range leaky {
				add("leaky_header:" + name)
			}
		}
	}
	if status, ok := results["status"].(map[string]any); ok {
		for path, info := range status {
			if m, ok := info.(map[string]any); ok {
				if exposed, _ := m["exposed"].(bool); exposed {
					add("exposed_path:" + path)
				}
			}
		}
	}
	for _, check := range []string{"sqli", "sqli_boolean", "xss", "path_traversal"} {
		if section, ok := results[check].(map[string]any); ok {
			if vuln, _ := section["vulnerable"].(bool); vuln {
				if list, ok := section["findings"].([]any); ok {
					for _, item := range list {
						param := "?"
						if m, ok := item.(map[string]any); ok {
							if p, ok := m["parameter"].(string); ok {
								param = p
							}
						}
						add(check + ":" + param)
					}
				}
			}
		}
	}
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
						if om, ok := item.(map[string]any); ok {
							add("open_port:" + ip + ":" + portStr(om["port"]))
						}
					}
				}
			}
		}
	}
	if cs, ok := results["cookies"].(map[string]any); ok {
		for _, cat := range []string{"missing_secure", "missing_httponly", "missing_samesite"} {
			for _, name := range strList(cs[cat]) {
				add("cookie_" + cat + ":" + name)
			}
		}
	}
	if co, ok := results["cors"].(map[string]any); ok {
		if bad, _ := co["misconfigured"].(bool); bad {
			add("cors_misconfig")
		}
	}
	if ms, ok := results["methods"].(map[string]any); ok {
		for _, m := range strList(ms["risky"]) {
			add("risky_method:" + m)
		}
	}
	if st, ok := results["security_txt"].(map[string]any); ok {
		if present, _ := st["present"].(bool); !present {
			add("missing_security_txt")
		}
	}
	if ss, ok := results["secrets"].(map[string]any); ok {
		if list, ok := ss["findings"].([]any); ok {
			for _, item := range list {
				if m, ok := item.(map[string]any); ok {
					kind, _ := m["kind"].(string)
					loc, _ := m["location"].(string)
					add("exposed_secret:" + kind + ":" + loc)
				}
			}
		}
	}

	out := make([]string, 0, len(set))
	for s := range set {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

// Result is the added/removed finding keys between two scans.
type Result struct {
	Added   []string `json:"added"`
	Removed []string `json:"removed"`
}

// Diff compares old → new signatures. Both lists are sorted and non-nil so
// the dashboard can render them directly.
func Diff(oldResults, newResults map[string]any) Result {
	oldSet := map[string]bool{}
	for _, s := range Signatures(oldResults) {
		oldSet[s] = true
	}
	newSet := map[string]bool{}
	for _, s := range Signatures(newResults) {
		newSet[s] = true
	}
	res := Result{Added: []string{}, Removed: []string{}}
	for s := range newSet {
		if !oldSet[s] {
			res.Added = append(res.Added, s)
		}
	}
	for s := range oldSet {
		if !newSet[s] {
			res.Removed = append(res.Removed, s)
		}
	}
	sort.Strings(res.Added)
	sort.Strings(res.Removed)
	return res
}
