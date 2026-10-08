package active

import (
	_ "embed"
	"strings"
	"sync"
	"sync/atomic"

	"fortify-go/internal/safe"
)

//go:embed config/traversal_payloads.txt
var traversalPayloadsRaw string

//go:embed config/traversal_signatures.txt
var traversalSigsRaw string

var traversalPayloads = loadLines(traversalPayloadsRaw)
var traversalSigs = lowerLines(loadLines(traversalSigsRaw))

type TraversalFinding struct {
	Parameter         string `json:"parameter"`
	Payload           string `json:"payload"`
	MatchedSignature  string `json:"matched_signature"`
}

type TraversalResult struct {
	Vulnerable   bool               `json:"vulnerable"`
	Findings     []TraversalFinding `json:"findings"`
	RequestsMade int64              `json:"requests_made"`
	Errors       int64              `json:"errors"`
}

func findFileSig(bodyLower string) string {
	for _, sig := range traversalSigs {
		if strings.Contains(bodyLower, sig) {
			return sig
		}
	}
	return ""
}

// ScanPathTraversal flags a parameter when a ../ payload leaks system-file
// contents (e.g. /etc/passwd markers). Parity with Python path_traversal.py.
func ScanPathTraversal(rawURL string) TraversalResult {
	params := keysOf(InjectPayload(rawURL, firstOr(traversalPayloads, "../")))
	var findings []TraversalFinding
	var mu sync.Mutex
	var made, errs atomic.Int64
	var wg sync.WaitGroup
	for _, param := range params {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			defer safe.Recover() // a panicking param is skipped, rest continue
			for _, payload := range traversalPayloads {
				injected := InjectPayload(rawURL, payload)[p]
				body, err := getBodyLower(injected)
				if err != nil {
					errs.Add(1)
					continue
				}
				made.Add(1)
				if matched := findFileSig(body); matched != "" {
					mu.Lock()
					findings = append(findings, TraversalFinding{Parameter: p, Payload: payload, MatchedSignature: matched})
					mu.Unlock()
					break
				}
			}
		}(param)
	}
	wg.Wait()
	if findings == nil {
		findings = []TraversalFinding{}
	}
	return TraversalResult{Vulnerable: len(findings) > 0, Findings: findings, RequestsMade: made.Load(), Errors: errs.Load()}
}
