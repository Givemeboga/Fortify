package active

import (
	_ "embed"
	"strings"
	"sync"
	"sync/atomic"

	"crypto/rand"
	"encoding/hex"

	"fortify-go/internal/scanner/httpclient"

	"io"
)

//go:embed config/xss_payloads.txt
var xssPayloadsRaw string

type XSSFinding struct {
	Parameter string `json:"parameter"`
	Payload   string `json:"payload"`
	URL       string `json:"url"`
}

type XSSResult struct {
	Vulnerable   bool         `json:"vulnerable"`
	Findings     []XSSFinding `json:"findings"`
	RequestsMade int64        `json:"requests_made"`
	Errors       int64        `json:"errors"`
}

// ScanXSS injects a unique per-scan token and flags a parameter only when the
// payload is reflected UNESCAPED (exact match) — escaped reflections are
// cleared and pre-existing page content can't false-positive. Parity with
// Python xss.py.
func ScanXSS(rawURL string) XSSResult {
	var rb [8]byte
	_, _ = rand.Read(rb[:])
	token := "fortify-" + hex.EncodeToString(rb[:])

	var payloads []string
	for _, p := range loadLines(xssPayloadsRaw) {
		payloads = append(payloads, strings.ReplaceAll(p, "{TOKEN}", token))
	}
	params := keysOf(InjectPayload(rawURL, firstOr(payloads, "1")))

	var findings []XSSFinding
	var mu sync.Mutex
	var made, errs atomic.Int64
	var wg sync.WaitGroup
	for _, param := range params {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			for _, payload := range payloads {
				injected := InjectPayload(rawURL, payload)[p]
				resp, err := httpclient.Shared.Get(injected)
				if err != nil {
					errs.Add(1)
					continue
				}
				body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
				resp.Body.Close()
				made.Add(1)
				if strings.Contains(string(body), payload) {
					mu.Lock()
					findings = append(findings, XSSFinding{Parameter: p, Payload: payload, URL: injected})
					mu.Unlock()
					break
				}
			}
		}(param)
	}
	wg.Wait()
	if findings == nil {
		findings = []XSSFinding{}
	}
	return XSSResult{Vulnerable: len(findings) > 0, Findings: findings, RequestsMade: made.Load(), Errors: errs.Load()}
}
