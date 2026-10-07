package active

import "strings"

// loadLines parses an embedded wordlist: strips blanks and # comments.
func loadLines(raw string) []string {
	var out []string
	for _, line := range strings.Split(raw, "\n") {
		l := strings.TrimSpace(line)
		if l == "" || strings.HasPrefix(l, "#") {
			continue
		}
		out = append(out, l)
	}
	return out
}

func lowerLines(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		out = append(out, strings.ToLower(s))
	}
	return out
}

func keysOf(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func firstOr(in []string, fallback string) string {
	if len(in) > 0 {
		return in[0]
	}
	return fallback
}
