package passive

import (
	"net/http"
	"strings"

	"fortify-go/internal/scanner/httpclient"
)

// CORSResult reports whether the app trusts an attacker origin. The dangerous
// combos are a reflected Origin (or *) TOGETHER WITH credentials — that lets
// any malicious site read authenticated responses.
type CORSResult struct {
	TestedOrigin     string `json:"tested_origin"`
	AllowOrigin      string `json:"allow_origin"`
	AllowCredentials bool   `json:"allow_credentials"`
	Misconfigured    bool   `json:"misconfigured"`
	Detail           string `json:"detail"`
}

const corsProbeOrigin = "https://evil-fortify-test.example"

// ScanCORS replays the request with an attacker Origin and inspects the
// trust decision. No state is changed — pure header inspection.
func ScanCORS(rawURL string) CORSResult {
	res := CORSResult{TestedOrigin: corsProbeOrigin}
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		res.Detail = "request build failed"
		return res
	}
	req.Header.Set("Origin", corsProbeOrigin)
	resp, err := httpclient.Shared.Do(req)
	if err != nil {
		res.Detail = "request failed"
		return res
	}
	defer resp.Body.Close()

	res.AllowOrigin = resp.Header.Get("Access-Control-Allow-Origin")
	res.AllowCredentials = strings.EqualFold(resp.Header.Get("Access-Control-Allow-Credentials"), "true")

	switch {
	case res.AllowOrigin == corsProbeOrigin && res.AllowCredentials:
		res.Misconfigured = true
		res.Detail = "Origin is reflected and credentials are allowed — any site can read authenticated responses"
	case res.AllowOrigin == "*" && res.AllowCredentials:
		res.Misconfigured = true
		res.Detail = "Wildcard origin with credentials allowed — any site can read authenticated responses"
	case res.AllowOrigin == corsProbeOrigin:
		res.Detail = "Origin is reflected, but credentials are not allowed — limited impact"
	case res.AllowOrigin == "*":
		res.Detail = "Wildcard origin without credentials — public API posture, limited impact"
	default:
		res.Detail = "Origin not trusted"
	}
	return res
}
