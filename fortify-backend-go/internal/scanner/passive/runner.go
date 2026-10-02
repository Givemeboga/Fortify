package passive

import (
	"net/url"
)

// RunPassiveScan mirrors Python scanner/passive/runner.py, extended with the
// Go-only checks (cookies, CORS, methods, security.txt, secrets, ports live
// at the API layer). Every probe is independent I/O, so all run concurrently;
// the result keys match what the dashboard and analyzer consume.
func RunPassiveScan(rawURL string) map[string]any {
	u, _ := url.Parse(rawURL)
	host := ""
	if u != nil {
		host = u.Hostname()
	}

	tlsCh := make(chan TLSResult, 1)
	headersCh := make(chan HeadersResult, 1)
	pathsCh := make(chan map[string]PathResult, 1)
	cookiesCh := make(chan CookiesResult, 1)
	corsCh := make(chan CORSResult, 1)
	methodsCh := make(chan MethodsResult, 1)
	sectxtCh := make(chan SecurityTxtResult, 1)
	secretsCh := make(chan SecretsResult, 1)

	go func() { tlsCh <- ScanTLS(host) }()
	go func() { headersCh <- ScanHeaders(rawURL) }()
	go func() { pathsCh <- ScanPaths(rawURL) }()
	go func() { cookiesCh <- ScanCookies(rawURL) }()
	go func() { corsCh <- ScanCORS(rawURL) }()
	go func() { methodsCh <- ScanMethods(rawURL) }()
	go func() { sectxtCh <- ScanSecurityTxt(rawURL) }()
	go func() { secretsCh <- ScanSecrets(rawURL) }()

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
