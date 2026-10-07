package passive

import (
	"net/http"

	"fortify-go/internal/scanner/httpclient"
)

// CookieIssue is one Set-Cookie with its parsed protection flags.
type CookieIssue struct {
	Name     string `json:"name"`
	Secure   bool   `json:"secure"`
	HttpOnly bool   `json:"http_only"`
	SameSite string `json:"same_site"` // Lax | Strict | None | "" (unset)
}

// CookiesResult groups cookies missing each flag so the analyzer can file
// one finding per category instead of flooding per cookie.
type CookiesResult struct {
	Cookies         []CookieIssue `json:"cookies"`
	MissingSecure   []string      `json:"missing_secure"`
	MissingHttpOnly []string      `json:"missing_httponly"`
	MissingSameSite []string      `json:"missing_samesite"`
}

// ScanCookies flags session/data cookies readable over HTTP (no Secure),
// readable from JS (no HttpOnly → XSS theft), or sent cross-site without
// an explicit SameSite policy (CSRF). Parity idea: header checks, but for
// the cookies the app actually sets.
func ScanCookies(rawURL string) CookiesResult {
	res := CookiesResult{}
	resp, err := httpclient.Shared.Get(rawURL)
	if err != nil {
		return res
	}
	defer resp.Body.Close()

	for _, c := range resp.Cookies() {
		sameSite := ""
		switch c.SameSite {
		case http.SameSiteLaxMode:
			sameSite = "Lax"
		case http.SameSiteStrictMode:
			sameSite = "Strict"
		case http.SameSiteNoneMode:
			sameSite = "None"
		}
		res.Cookies = append(res.Cookies, CookieIssue{
			Name: c.Name, Secure: c.Secure, HttpOnly: c.HttpOnly, SameSite: sameSite,
		})
		if !c.Secure {
			res.MissingSecure = append(res.MissingSecure, c.Name)
		}
		if !c.HttpOnly {
			res.MissingHttpOnly = append(res.MissingHttpOnly, c.Name)
		}
		if sameSite == "" {
			res.MissingSameSite = append(res.MissingSameSite, c.Name)
		}
	}
	if res.Cookies == nil {
		res.Cookies = []CookieIssue{}
	}
	if res.MissingSecure == nil {
		res.MissingSecure = []string{}
	}
	if res.MissingHttpOnly == nil {
		res.MissingHttpOnly = []string{}
	}
	if res.MissingSameSite == nil {
		res.MissingSameSite = []string{}
	}
	return res
}
