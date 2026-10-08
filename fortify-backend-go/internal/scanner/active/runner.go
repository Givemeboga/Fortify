package active

import "fortify-go/internal/safe"

// RunActiveScan mirrors Python scanner/active/runner.py. The four injection
// families are independent I/O workloads, so they run concurrently — the
// Python version ran them one after another. Each family is panic-contained.
func RunActiveScan(rawURL string) map[string]any {
	type kv struct {
		k string
		v any
	}
	ch := make(chan kv, 4)
	go safe.Send(ch, kv{"sqli", SQLiResult{Findings: []SQLiFinding{}}},
		func() kv { return kv{"sqli", ScanSQLi(rawURL)} })
	go safe.Send(ch, kv{"sqli_boolean", SQLiBooleanResult{Findings: []SQLiBooleanFinding{}}},
		func() kv { return kv{"sqli_boolean", ScanSQLiBoolean(rawURL)} })
	go safe.Send(ch, kv{"xss", XSSResult{Findings: []XSSFinding{}}},
		func() kv { return kv{"xss", ScanXSS(rawURL)} })
	go safe.Send(ch, kv{"path_traversal", TraversalResult{Findings: []TraversalFinding{}}},
		func() kv { return kv{"path_traversal", ScanPathTraversal(rawURL)} })

	out := make(map[string]any, 4)
	for range 4 {
		item := <-ch
		out[item.k] = item.v
	}
	return out
}
