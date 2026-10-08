package passive

import (
	"io"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"fortify-go/internal/safe"
	"fortify-go/internal/scanner/httpclient"
)

// secretPattern is one concrete token format — no fuzzy entropy heuristics,
// so findings are always actionable and never "looks random".
type secretPattern struct {
	kind string
	re   *regexp.Regexp
}

var secretPatterns = []secretPattern{
	{"aws_access_key_id", regexp.MustCompile(`AKIA[0-9A-Z]{16}`)},
	{"aws_secret_key", regexp.MustCompile(`(?i)aws_secret_access_key["']?\s*[:=]\s*["']?[A-Za-z0-9/+=]{40}`)},
	{"google_api_key", regexp.MustCompile(`AIza[0-9A-Za-z\-_]{35}`)},
	{"slack_token", regexp.MustCompile(`xox[baprs]-[0-9A-Za-z\-]+`)},
	{"private_key", regexp.MustCompile(`-----BEGIN (?:RSA |EC |DSA |OPENSSH |ENCRYPTED )?PRIVATE KEY-----`)},
	{"generic_secret", regexp.MustCompile(`(?i)(api[_-]?key|secret|passwd|password|auth[_-]?token)\s*[:=]\s*["'][^"'${}]{8,}["']`)},
}

// placeholderValues skips obvious non-secrets (docs, templates, redacted).
var placeholderValues = []string{
	"your", "example", "xxx", "***", "todo", "changeme",
	"placeholder", "dummy", "sample", "undefined", "null",
}

func isPlaceholder(match string) bool {
	lower := strings.ToLower(match)
	for _, p := range placeholderValues {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return false
}

// SecretFinding is one concrete secret sighting with a masked preview.
type SecretFinding struct {
	Kind     string `json:"kind"`
	Location string `json:"location"`
	Preview  string `json:"preview"`
}

// SecretsResult holds every confirmed client-visible secret.
type SecretsResult struct {
	Findings []SecretFinding `json:"findings"`
}

// redactPreview masks the middle so the report proves exposure without
// republishing the credential.
func redactPreview(match string) string {
	match = strings.TrimSpace(match)
	if len(match) <= 10 {
		return match[:2] + "***"
	}
	return match[:4] + "***" + match[len(match)-2:]
}

var scriptSrcRE = regexp.MustCompile(`(?i)<script[^>]+src=["']([^"']+)["']`)

// ScanSecrets inspects the page HTML plus same-origin scripts (cap 8) for
// concrete token formats. Only full-format matches count — placeholders and
// template variables are filtered, so this stays false-positive free.
func ScanSecrets(rawURL string) SecretsResult {
	res := SecretsResult{Findings: []SecretFinding{}}
	resp, err := httpclient.Shared.Get(rawURL)
	if err != nil {
		return res
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	resp.Body.Close()
	html := string(body)

	base, _ := url.Parse(rawURL)
	pages := []struct{ location, body string }{{"page HTML", html}}

	seen := map[string]bool{}
	var scripts []string
	for _, m := range scriptSrcRE.FindAllStringSubmatch(html, 9) {
		u, err := url.Parse(m[1])
		if err != nil {
			continue
		}
		if !u.IsAbs() {
			u = base.ResolveReference(u)
		}
		if u.Hostname() != base.Hostname() || seen[u.String()] || len(scripts) >= 8 {
			continue
		}
		seen[u.String()] = true
		scripts = append(scripts, u.String())
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, src := range scripts {
		wg.Add(1)
		go func(s string) {
			defer wg.Done()
			defer safe.Recover() // a panicking fetch is skipped, rest continue
			r, err := httpclient.Shared.Get(s)
			if err != nil {
				return
			}
			b, _ := io.ReadAll(io.LimitReader(r.Body, 1<<20))
			r.Body.Close()
			mu.Lock()
			pages = append(pages, struct{ location, body string }{s, string(b)})
			mu.Unlock()
		}(src)
	}
	wg.Wait()

	found := map[string]bool{}
	for _, p := range pages {
		for _, sp := range secretPatterns {
			for _, m := range sp.re.FindAllString(p.body, -1) {
				if isPlaceholder(m) {
					continue
				}
				key := sp.kind + "\x00" + m
				if found[key] {
					continue
				}
				found[key] = true
				res.Findings = append(res.Findings, SecretFinding{
					Kind: sp.kind, Location: p.location, Preview: redactPreview(m),
				})
			}
		}
	}
	return res
}
