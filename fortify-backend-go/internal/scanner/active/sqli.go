package active

import (
	_ "embed"
	"io"
	"strings"
	"sync"
	"sync/atomic"

	"fortify-go/internal/scanner/httpclient"
)

//go:embed config/sqli_payloads.txt
var sqliPayloadsRaw string

//go:embed config/sql_errors.txt
var sqlErrorsRaw string

var sqliPayloads = loadLines(sqliPayloadsRaw)
var sqlErrors = lowerLines(loadLines(sqlErrorsRaw))

// Boolean-based blind pairs (TRUE vs FALSE condition). A parameter is flagged
// when the two responses differ meaningfully — catches injections that leak
// no error text. Parity with Python BOOLEAN_PAYLOADS.
var booleanPairs = [][2]string{
	{"1 AND 1=1", "1 AND 1=2"},
	{"1' AND '1'='1", "1' AND '1'='2"},
}

type SQLiFinding struct {
	Parameter    string `json:"parameter"`
	Payload      string `json:"payload,omitempty"`
	MatchedError string `json:"matched_error,omitempty"`
}

type SQLiBooleanFinding struct {
	Parameter    string `json:"parameter"`
	TruePayload  string `json:"true_payload"`
	FalsePayload string `json:"false_payload"`
}

type SQLiResult struct {
	Vulnerable   bool          `json:"vulnerable"`
	Findings     []SQLiFinding `json:"findings"`
	RequestsMade int64         `json:"requests_made"`
	Errors       int64         `json:"errors"`
}

type SQLiBooleanResult struct {
	Vulnerable   bool                 `json:"vulnerable"`
	Findings     []SQLiBooleanFinding `json:"findings"`
	RequestsMade int64                `json:"requests_made"`
	Errors       int64                `json:"errors"`
}

func findSQLError(bodyLower string) string {
	for _, sig := range sqlErrors {
		if strings.Contains(bodyLower, sig) {
			return sig
		}
	}
	return ""
}

func getBodyLower(url string) (string, error) {
	resp, err := httpclient.Shared.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	return strings.ToLower(string(b)), nil
}

// ScanSQLi: error-based SQLi. Params are probed concurrently (one goroutine
// per parameter; payloads sequential within a param, first hit wins — same
// break-on-confirm semantics as Python). Counters are atomic so a failed
// scan is never mistaken for a clean one.
func ScanSQLi(rawURL string) SQLiResult {
	params := keysOf(InjectPayload(rawURL, firstOr(sqliPayloads, "1")))
	var findings []SQLiFinding
	var mu sync.Mutex
	var made, errs atomic.Int64
	var wg sync.WaitGroup
	for _, param := range params {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			for _, payload := range sqliPayloads {
				injected := InjectPayload(rawURL, payload)[p]
				body, err := getBodyLower(injected)
				if err != nil {
					errs.Add(1)
					continue
				}
				made.Add(1)
				if matched := findSQLError(body); matched != "" {
					mu.Lock()
					findings = append(findings, SQLiFinding{Parameter: p, Payload: payload, MatchedError: matched})
					mu.Unlock()
					break // confirmed for this param — move on
				}
			}
		}(param)
	}
	wg.Wait()
	if findings == nil {
		findings = []SQLiFinding{}
	}
	return SQLiResult{Vulnerable: len(findings) > 0, Findings: findings, RequestsMade: made.Load(), Errors: errs.Load()}
}

// ScanSQLiBoolean: boolean-based blind SQLi (TRUE vs FALSE response diff).
func ScanSQLiBoolean(rawURL string) SQLiBooleanResult {
	params := keysOf(InjectPayload(rawURL, "1"))
	var findings []SQLiBooleanFinding
	var mu sync.Mutex
	var made, errs atomic.Int64
	var wg sync.WaitGroup
	for _, param := range params {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			for _, pair := range booleanPairs {
				trueBody, err1 := getBodyLower(InjectPayload(rawURL, pair[0])[p])
				falseBody, err2 := getBodyLower(InjectPayload(rawURL, pair[1])[p])
				if err1 != nil || err2 != nil {
					errs.Add(1)
					continue
				}
				made.Add(2)
				if len(trueBody) != len(falseBody) { // _similar tolerance 0
					mu.Lock()
					findings = append(findings, SQLiBooleanFinding{Parameter: p, TruePayload: pair[0], FalsePayload: pair[1]})
					mu.Unlock()
					break
				}
			}
		}(param)
	}
	wg.Wait()
	if findings == nil {
		findings = []SQLiBooleanFinding{}
	}
	return SQLiBooleanResult{Vulnerable: len(findings) > 0, Findings: findings, RequestsMade: made.Load(), Errors: errs.Load()}
}
