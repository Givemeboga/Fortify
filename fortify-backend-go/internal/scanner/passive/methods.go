package passive

import (
	"net/http"
	"strings"

	"fortify-go/internal/scanner/httpclient"
)

// RiskyMethods enable request smuggling / cross-protocol tricks (TRACE for
// XST cookie theft) or unintended state changes (PUT/DELETE) on the origin.
var riskyMethods = []string{"TRACE", "TRACK", "PUT", "DELETE", "CONNECT"}

// MethodsResult lists the methods the origin advertises plus the risky subset.
type MethodsResult struct {
	Allowed []string `json:"allowed"`
	Risky   []string `json:"risky"`
}

// ScanMethods asks for OPTIONS and additionally probes TRACE directly —
// some servers honor TRACE without advertising it in Allow.
func ScanMethods(rawURL string) MethodsResult {
	res := MethodsResult{Allowed: []string{}, Risky: []string{}}
	seen := map[string]bool{}

	add := func(m string) {
		m = strings.ToUpper(strings.TrimSpace(m))
		if m == "" || seen[m] {
			return
		}
		seen[m] = true
		res.Allowed = append(res.Allowed, m)
	} 

	req, err := http.NewRequest("OPTIONS", rawURL, nil)
	if err == nil {
		if resp, err := httpclient.Shared.Do(req); err == nil {
			for _, m := range strings.Split(resp.Header.Get("Allow"), ",") {
				add(m)
			}
			// Some stacks advertise via Access-Control-Allow-Methods instead.
			for _, m := range strings.Split(resp.Header.Get("Access-Control-Allow-Methods"), ",") {
				add(m)
			}
			resp.Body.Close()
		}
	}

	// Direct TRACE probe: 2xx/3xx means the verb is live regardless of Allow.
	if treq, err := http.NewRequest("TRACE", rawURL, nil); err == nil {
		if tresp, err := httpclient.Shared.Do(treq); err == nil {
			if tresp.StatusCode < 400 {
				add("TRACE")
			}
			tresp.Body.Close()
		}
	}

	for _, m := range res.Allowed {
		for _, r := range riskyMethods {
			if m == r {
				res.Risky = append(res.Risky, m)
				break
			}
		}
	}
	return res
}
