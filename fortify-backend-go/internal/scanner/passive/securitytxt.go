package passive

import (
	"io"
	"strings"

	"fortify-go/internal/scanner/httpclient"
)

// SecurityTxtResult reports the RFC 9116 vulnerability-disclosure file.
type SecurityTxtResult struct {
	Present      bool   `json:"present"`
	StatusCode   *int   `json:"status_code"`
	ContentType  string `json:"content_type"`
	HasContact   bool   `json:"has_contact"`
}

// ScanSecurityTxt fetches /.well-known/security.txt. Present means HTTP 200
// with a Contact: field — anything else leaves researchers with no channel.
func ScanSecurityTxt(rawURL string) SecurityTxtResult {
	res := SecurityTxtResult{}
	base := strings.TrimRight(rawURL, "/")
	// Strip path/query — security.txt lives at the origin root.
	if i := strings.Index(base, "://"); i >= 0 {
		if j := strings.Index(base[i+3:], "/"); j >= 0 {
			base = base[:i+3+j]
		}
	}
	resp, err := httpclient.Shared.Get(base + "/.well-known/security.txt")
	if err != nil {
		return res
	}
	defer resp.Body.Close()
	code := resp.StatusCode
	res.StatusCode = &code
	res.ContentType = resp.Header.Get("Content-Type")
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
	res.HasContact = strings.Contains(strings.ToLower(string(body)), "contact:")
	res.Present = resp.StatusCode == 200 && res.HasContact
	return res
}
