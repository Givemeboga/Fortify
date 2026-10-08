package passive

import (
	_ "embed"
	"io"
	"strings"
	"sync"

	"fortify-go/internal/safe"
	"fortify-go/internal/scanner/httpclient"
)

//go:embed config/paths.txt
var pathsRaw string

//go:embed config/benign_paths.txt
var benignRaw string

// Listing signatures: a 200 HTML body is a real exposure only with one of
// these fingerprints (open listing or raw server-side source). Plain app
// pages default to safe. Parity with Python status.py / issues #15/#16.
var listingSignatures = []string{"Index of /", "Parent Directory", "<?php", "-----BEGIN"}

type PathResult struct {
	StatusCode *int `json:"status_code"`
	Exposed    bool `json:"exposed"`
}

func loadWordlist(raw string) []string {
	var out []string
	seen := map[string]bool{}
	for _, line := range strings.Split(raw, "\n") {
		l := strings.TrimSpace(line)
		if l == "" || strings.HasPrefix(l, "#") || seen[l] {
			continue
		}
		seen[l] = true
		out = append(out, l)
	}
	return out
}

func probe(url string) (status *int, body, ctype string) {
	resp, err := httpclient.Shared.Get(url)
	if err != nil {
		return nil, "", ""
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1 MiB cap
	s := resp.StatusCode
	return &s, string(b), resp.Header.Get("Content-Type")
}

func isExposed(status *int, body, ctype string) bool {
	if status == nil || *status != 200 {
		return false
	}
	if !strings.Contains(strings.ToLower(ctype), "text/html") {
		// Non-HTML file leak (text/plain, sql, octet-stream, ...) — real file.
		return true
	}
	for _, sig := range listingSignatures {
		if strings.Contains(body, sig) {
			return true
		}
	}
	return false
}

// ScanPaths probes every sensitive path concurrently (bounded worker pool).
// The Python original probed sequentially; with ~10 paths x 10s worst-case
// timeouts this is the single biggest scan-time win in the Go port.
func ScanPaths(base string) map[string]PathResult {
	base = strings.TrimRight(base, "/")
	paths := loadWordlist(pathsRaw)
	benign := map[string]bool{}
	for _, p := range loadWordlist(benignRaw) {
		benign[p] = true
	}

	results := make(map[string]PathResult, len(paths))
	var mu sync.Mutex
	sem := make(chan struct{}, 10) // max 10 in-flight probes
	var wg sync.WaitGroup
	for _, p := range paths {
		if benign[p] {
			results[p] = PathResult{StatusCode: intp(200), Exposed: false}
			continue
		}
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			safe.Do(func() {
				st, body, ctype := probe(base + path)
				mu.Lock()
				results[path] = PathResult{StatusCode: st, Exposed: isExposed(st, body, ctype)}
				mu.Unlock()
			})
		}(p)
	}
	wg.Wait()
	return results
}

func intp(i int) *int { return &i }
