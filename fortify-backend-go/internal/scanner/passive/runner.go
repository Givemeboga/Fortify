package passive

import (
	"net/url"
	"strconv"

	"fortify-go/internal/safe"
)

// RunPassiveScan mirrors Python scanner/passive/runner.py, extended with the
// Go-only checks (cookies, CORS, methods, security.txt, secrets, ports live
// at the API layer). Every probe is independent I/O, so all run concurrently;
// the result keys match what the dashboard and analyzer consume.
func RunPassiveScan(rawURL string) map[string]any {
	u, _ := url.Parse(rawURL)
	host := ""
	port := 443
	if u != nil {
		host = u.Hostname()
		// Honor explicit ports (https://host:8443) instead of always 443.
		if p := u.Port(); p != "" {
			if n, err := strconv.Atoi(p); err == nil && n > 0 {
				port = n
			}
		}
	}

	tlsCh := make(chan TLSResult, 1)
	headersCh := make(chan HeadersResult, 1)
	pathsCh := make(chan map[string]PathResult, 1)
	cookiesCh := make(chan CookiesResult, 1)
	corsCh := make(chan CORSResult, 1)
	methodsCh := make(chan MethodsResult, 1)
	sectxtCh := make(chan SecurityTxtResult, 1)
	secretsCh := make(chan SecretsResult, 1)

	// Each probe is panic-contained: a crashing check reports its zero value
	// instead of taking down the server (see Blocker 4 review).
	go safe.Send(tlsCh, TLSResult{}, func() TLSResult { return ScanTLS(host, port) })
	go safe.Send(headersCh, HeadersResult{}, func() HeadersResult { return ScanHeaders(rawURL) })
	go safe.Send(pathsCh, map[string]PathResult{}, func() map[string]PathResult { return ScanPaths(rawURL) })
	go safe.Send(cookiesCh, CookiesResult{}, func() CookiesResult { return ScanCookies(rawURL) })
	go safe.Send(corsCh, CORSResult{}, func() CORSResult { return ScanCORS(rawURL) })
	go safe.Send(methodsCh, MethodsResult{}, func() MethodsResult { return ScanMethods(rawURL) })
	go safe.Send(sectxtCh, SecurityTxtResult{}, func() SecurityTxtResult { return ScanSecurityTxt(rawURL) })
	go safe.Send(secretsCh, SecretsResult{}, func() SecretsResult { return ScanSecrets(rawURL) })

	return map[string]any{
		"tls":          <-tlsCh,
		"headers":      <-headersCh,
		"status":       <-pathsCh,
		"cookies":      <-cookiesCh,
		"cors":         <-corsCh,
		"methods":      <-methodsCh,
		"security_txt": <-sectxtCh,
		"secrets":      <-secretsCh,
	}
}
