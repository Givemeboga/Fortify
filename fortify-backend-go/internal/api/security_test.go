package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthOpenWithoutToken(t *testing.T) {
	t.Setenv("FORTIFY_API_TOKEN", "")
	h := auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/scans", nil))
	if rec.Code != http.StatusTeapot {
		t.Fatalf("status = %d, want passthrough", rec.Code)
	}
}

func TestAuthEnforcedWithToken(t *testing.T) {
	t.Setenv("FORTIFY_API_TOKEN", "s3cret")
	h := auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/scans", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous status = %d, want 401", rec.Code)
	}

	rec = httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/scans", nil)
	req.Header.Set("Authorization", "Bearer s3cret")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusTeapot {
		t.Fatalf("authed status = %d, want passthrough", rec.Code)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/scans", nil)
	req.Header.Set("Authorization", "Bearer wrong")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("wrong-key status = %d, want 401", rec.Code)
	}
}

func TestAuthSkipsHealthz(t *testing.T) {
	t.Setenv("FORTIFY_API_TOKEN", "s3cret")
	h := auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/healthz", nil))
	if rec.Code != http.StatusTeapot {
		t.Fatalf("healthz status = %d, want passthrough", rec.Code)
	}
}

func TestCheckPublicHostLiterals(t *testing.T) {
	for _, host := range []string{"127.0.0.1", "::1", "10.0.0.5", "192.168.1.1", "172.16.0.1", "169.254.169.254", "0.0.0.0", "224.0.0.1", "fd00::1", "fe80::1"} {
		if err := checkPublicHost(host); err == nil {
			t.Fatalf("%s should be rejected", host)
		}
	}
	for _, host := range []string{"8.8.8.8", "1.1.1.1", "93.184.216.34", "2606:4700:4700::1111"} {
		if err := checkPublicHost(host); err != nil {
			t.Fatalf("%s should pass, got %v", host, err)
		}
	}
}

func TestCheckPublicHostOptOut(t *testing.T) {
	t.Setenv("FORTIFY_ALLOW_PRIVATE_IPS", "1")
	if err := checkPublicHost("192.168.1.1"); err != nil {
		t.Fatalf("opt-out should allow private IPs, got %v", err)
	}
}
