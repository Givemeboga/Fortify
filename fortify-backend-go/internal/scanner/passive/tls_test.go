package passive

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

// The httptest TLS server uses a self-signed cert: the verified handshake
// fails, so the scanner must fall back to inspecting the real certificate
// instead of returning all-nulls.
func TestScanTLSSelfSigned(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()
	u, _ := url.Parse(srv.URL)
	port, _ := strconv.Atoi(u.Port())

	res := ScanTLS(u.Hostname(), port)
	if res.TLSVersion == nil || *res.TLSVersion == "" {
		t.Fatal("expected a negotiated TLS version from inspection")
	}
	if res.CertValid {
		t.Fatal("self-signed cert must not validate")
	}
	if res.CertExpired == nil {
		t.Fatal("expected an expiry verdict from the real cert")
	}
	if res.VerificationError == nil || !strings.Contains(strings.ToLower(*res.VerificationError), "certificate") {
		t.Fatalf("expected a certificate verification error, got %+v", res.VerificationError)
	}
}

func TestScanTLSUnreachable(t *testing.T) {
	res := ScanTLS("127.0.0.1", 1) // port 1: nothing listens
	if res.TLSVersion != nil || res.CertExpired != nil || res.CertValid {
		t.Fatalf("unreachable port should be all-null, got %+v", res)
	}
}

func TestScanTLSUsesVerifiedPath(t *testing.T) {
	// Sanity: a verified handshake still validates. example.com is public;
	// skip when offline.
	conn, err := tls.Dial("tcp", "example.com:443", &tls.Config{MinVersion: tls.VersionTLS10})
	if err != nil {
		t.Skipf("offline, skipping: %v", err)
	}
	conn.Close()
	res := ScanTLS("example.com", 443)
	if !res.CertValid || res.TLSVersion == nil {
		t.Fatalf("public cert should validate: %+v", res)
	}
}
