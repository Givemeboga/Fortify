package api

import (
	"crypto/subtle"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
)

var warnOnce sync.Once

// auth wraps the router with a shared-token check. Set FORTIFY_API_TOKEN and
// every route except /healthz requires `Authorization: Bearer <token>`.
// Unset (local dev default) leaves the API open and logs one warning.
func auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// /healthz stays open for load-balancer checks; OPTIONS preflights
		// carry no credentials by design, so they must pass too — otherwise
		// enabling the token would break every cross-origin call entirely.
		if r.URL.Path == "/healthz" || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}
		want := strings.TrimSpace(os.Getenv("FORTIFY_API_TOKEN"))
		if want == "" {
			warnOnce.Do(func() {
				log.Print("[auth] FORTIFY_API_TOKEN unset — API is open to anyone who can reach it")
			})
			next.ServeHTTP(w, r)
			return
		}
		got := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
		if got == "" || subtle.ConstantTimeCompare([]byte(got), []byte(want)) != 1 {
			writeDetail(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// allowedOrigins parses ALLOWED_ORIGINS (comma-separated, default the local
// dev dashboard). A literal "*" allows any origin (no credentials used).
func allowedOrigins() []string {
	if v := strings.TrimSpace(os.Getenv("ALLOWED_ORIGINS")); v != "" {
		var out []string
		for _, o := range strings.Split(v, ",") {
			if o = strings.TrimSpace(o); o != "" {
				out = append(out, o)
			}
		}
		if len(out) > 0 {
			return out
		}
	}
	return []string{"http://localhost:5173"}
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origins := allowedOrigins()
		origin := r.Header.Get("Origin")
		allow := ""
		for _, o := range origins {
			if o == "*" {
				allow = "*"
				break
			}
			if origin != "" && o == origin {
				allow = origin
				break
			}
		}
		if allow != "" {
			w.Header().Set("Access-Control-Allow-Origin", allow)
			w.Header().Set("Vary", "Origin")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// targetHost extracts the hostname from an already-validated target URL.
func targetHost(raw string) string {
	u, err := url.ParseRequestURI(strings.TrimSpace(raw))
	if err != nil {
		return ""
	}
	return u.Hostname()
}

func allowPrivateIPs() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("FORTIFY_ALLOW_PRIVATE_IPS")))
	return v == "1" || v == "true" || v == "yes"
}

// nonPublicReason reports why an IP must not be a scan target.
func nonPublicReason(ip net.IP) string {
	switch {
	case ip.IsUnspecified():
		return "unspecified address"
	case ip.IsLoopback():
		return "loopback address"
	case ip.IsLinkLocalUnicast():
		return "link-local address"
	case ip.IsLinkLocalMulticast(), ip.IsMulticast():
		return "multicast address"
	case ip.IsPrivate():
		return "private-network address"
	}
	return ""
}

// checkPublicHost resolves a hostname and rejects it when ANY address is
// non-public (loopback, RFC1918/ULA, link-local, multicast, …). This closes
// the SSRF hole where attacker.example resolves to 169.254.169.254 or a
// neighbour container — DNS is checked, not just the literal. Local labs can
// opt out with FORTIFY_ALLOW_PRIVATE_IPS=1.
func checkPublicHost(host string) error {
	if host == "" || allowPrivateIPs() {
		return nil
	}
	if ip := net.ParseIP(host); ip != nil {
		if reason := nonPublicReason(ip); reason != "" {
			return fmt.Errorf("target is a %s (%s)", reason, host)
		}
		return nil
	}
	addrs, err := net.LookupIP(host)
	if err != nil || len(addrs) == 0 {
		return fmt.Errorf("target did not resolve (%s)", host)
	}
	for _, a := range addrs {
		if reason := nonPublicReason(a); reason != "" {
			return fmt.Errorf("target resolves to a %s (%s)", reason, a.String())
		}
	}
	return nil
}
