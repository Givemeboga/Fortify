package active

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

// /stable reflects deterministically: TRUE vs FALSE payloads produce
// different bodies every time — the classic injectable shape.
// /token appends a growing nonce, so identical requests always differ —
// the dynamic-page shape that must NOT flag.
func booleanFixture(t *testing.T) *httptest.Server {
	var calls atomic.Int64
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := r.URL.Query().Get("q")
		switch r.URL.Path {
		case "/stable":
			if strings.Contains(v, "1=1") {
				fmt.Fprint(w, "welcome back, admin")
			} else {
				fmt.Fprint(w, "welcome back, guest — please log in to continue reading today")
			}
		case "/token":
			n := calls.Add(1)
			fmt.Fprintf(w, "page-%d csrf=%d-%s", n, n*7919, v)
		}
	}))
}

func TestScanSQLiBooleanStableFlags(t *testing.T) {
	srv := booleanFixture(t)
	defer srv.Close()
	res := ScanSQLiBoolean(srv.URL + "/stable?q=1")
	if !res.Vulnerable || len(res.Findings) != 1 || res.Findings[0].Parameter != "q" {
		t.Fatalf("stable TRUE/FALSE split should flag: %+v", res)
	}
}

func TestScanSQLiBooleanDynamicPageClean(t *testing.T) {
	srv := booleanFixture(t)
	defer srv.Close()
	res := ScanSQLiBoolean(srv.URL + "/token?q=1")
	if res.Vulnerable {
		t.Fatalf("self-varying page must not flag: %+v", res)
	}
}
