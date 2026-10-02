package active

// RunActiveScan mirrors Python scanner/active/runner.py. The four injection
// families are independent I/O workloads, so they run concurrently — the
// Python version ran them one after another.
func RunActiveScan(rawURL string) map[string]any {
	type kv struct {
		k string
		v any
	}
	ch := make(chan kv, 4)
	go func() { ch <- kv{"sqli", ScanSQLi(rawURL)} }()
	go func() { ch <- kv{"sqli_boolean", ScanSQLiBoolean(rawURL)} }()
	go func() { ch <- kv{"xss", ScanXSS(rawURL)} }()
	go func() { ch <- kv{"path_traversal", ScanPathTraversal(rawURL)} }()

	out := make(map[string]any, 4)
	for range 4 {
		item := <-ch
		out[item.k] = item.v
	}
	return out
}
