// Package httpclient holds the single shared HTTP client used by every
// scanner probe. Connection reuse (keep-alive pool) is the main reason the
// Go port is faster than the Python original, which built a new session per
// request: dozens of probes against one host now share a handful of TCP/TLS
// connections instead of re-handshaking each time.
package httpclient

import "net/http"
import "time"

var Shared = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
	},
}
