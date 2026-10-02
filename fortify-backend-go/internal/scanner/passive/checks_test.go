package passive

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func testServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/security.txt":
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("Contact: mailto:security@example.com\n"))
			return
		}
		if r.Method == "OPTIONS" {
			w.Header().Set("Allow", "GET, HEAD, OPTIONS, TRACE")
			w.WriteHeader(200)
			return
		}
		if r.Method == "TRACE" {
			w.WriteHeader(200)
			return
		}
		w.Header().Add("Set-Cookie", "session=abc123; Path=/")
		w.Header().Add("Set-Cookie", "prefs=dark; Path=/; Secure; HttpOnly; SameSite=Lax")
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><body><script>var key="AKIAIOSFODNN7XK9Q2M4";</script></body></html>`))
	}))
}

func TestScanCookies(t *testing.T) {
	srv := testServer()
	defer srv.Close()
	res := ScanCookies(srv.URL)
	if len(res.Cookies) != 2 {
		t.Fatalf("got %d cookies, want 2", len(res.Cookies))
	}
	if len(res.MissingSecure) != 1 || res.MissingSecure[0] != "session" {
		t.Fatalf("missing_secure = %v", res.MissingSecure)
	}
	if len(res.MissingHttpOnly) != 1 || res.MissingHttpOnly[0] != "session" {
		t.Fatalf("missing_httponly = %v", res.MissingHttpOnly)
	}
	if len(res.MissingSameSite) != 1 || res.MissingSameSite[0] != "session" {
		t.Fatalf("missing_samesite = %v", res.MissingSameSite)
	}
}

func TestScanCORS(t *testing.T) {
	srv := testServer()
	defer srv.Close()
	res := ScanCORS(srv.URL)
	if !res.Misconfigured {
		t.Fatal("reflected origin + credentials should be misconfigured")
	}
	if res.AllowOrigin != corsProbeOrigin {
		t.Fatalf("allow_origin = %q", res.AllowOrigin)
	}
}

func TestScanMethods(t *testing.T) {
	srv := testServer()
	defer srv.Close()
	res := ScanMethods(srv.URL)
	found := false
	for _, m := range res.Risky {
		if m == "TRACE" {
			found = true
		}
	}
	if !found {
		t.Fatalf("TRACE should be risky, allowed = %v", res.Allowed)
	}
}

func TestScanSecurityTxt(t *testing.T) {
	srv := testServer()
	defer srv.Close()
	res := ScanSecurityTxt(srv.URL + "/some/page?q=1")
	if !res.Present || !res.HasContact {
		t.Fatalf("security.txt should be present: %+v", res)
	}
}

func TestScanSecrets(t *testing.T) {
	srv := testServer()
	defer srv.Close()
	res := ScanSecrets(srv.URL)
	if len(res.Findings) != 1 {
		t.Fatalf("got %d findings, want 1: %+v", len(res.Findings), res.Findings)
	}
	f := res.Findings[0]
	if f.Kind != "aws_access_key_id" {
		t.Fatalf("kind = %q", f.Kind)
	}
	if f.Preview == "AKIAIOSFODNN7XK9Q2M4" || len(f.Preview) >= len("AKIAIOSFODNN7XK9Q2M4") {
		t.Fatalf("preview not redacted: %q", f.Preview)
	}
}

func TestScanSecretsPlaceholdersIgnored(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`var k = "YOUR_API_KEY_HERE"; var s = "AKIAIOSFODNN7EXAMPLE";`))
	}))
	defer srv.Close()
	res := ScanSecrets(srv.URL)
	if len(res.Findings) != 0 {
		t.Fatalf("placeholders should be ignored, got %+v", res.Findings)
	}
}
