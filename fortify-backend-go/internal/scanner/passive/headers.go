package passive

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"regexp"

	"fortify-go/internal/scanner/httpclient"
)

//go:embed config/headers.json
var headersJSON []byte

// VERSION_RE mirrors Python: pulls "software/version" out of leaky values,
// e.g. Server: nginx/1.18.0 -> {software: nginx, version: 1.18.0}.
var versionRE = regexp.MustCompile(`([A-Za-z][\w.\-]*)/(\d+(?:\.\d+)+)`)

type headersConfig struct {
	Defensive []string `json:"defensive"`
	Leaky     []string `json:"leaky"`
}

type VersionDisclosure struct {
	Header   string `json:"header"`
	Software string `json:"software"`
	Version  string `json:"version"`
}

type RedirectInfo struct {
	FinalURL   string   `json:"final_url"`
	Redirected bool     `json:"redirected"`
	Chain      []string `json:"chain"`
}

type HeadersResult struct {
	MissingHeaders     []string                     `json:"missing_headers"`
	PresentHeaders     []string                     `json:"present_headers"`
	LeakyHeaders       map[string]string            `json:"leaky_headers"`
	VersionDisclosures map[string]VersionDisclosure `json:"version_disclosures"`
	RedirectInfo       RedirectInfo                 `json:"redirect_info"`
}

// ScanHeaders mirrors Python scanner/passive/headers.py: missing vs present
// defensive headers, leaky header values, version disclosures, redirect chain.
// A per-request client with a redirect recorder is used so the chain matches
// Python's response.history; all other probes share httpclient.Shared.
func ScanHeaders(url string) HeadersResult {
	empty := HeadersResult{
		LeakyHeaders:       map[string]string{},
		VersionDisclosures: map[string]VersionDisclosure{},
		RedirectInfo:       RedirectInfo{Chain: []string{}},
	}
	var cfg headersConfig
	if err := json.Unmarshal(headersJSON, &cfg); err != nil {
		return empty
	}

	var chain []string
	client := &http.Client{
		Timeout:   httpclient.Shared.Timeout,
		Transport: httpclient.Shared.Transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			for _, r := range via {
				chain = append(chain, r.URL.String())
			}
			if len(via) >= 10 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}
	resp, err := client.Get(url)
	if err != nil {
		return empty
	}
	defer resp.Body.Close()

	res := HeadersResult{
		LeakyHeaders:       map[string]string{},
		VersionDisclosures: map[string]VersionDisclosure{},
	}
	for _, h := range cfg.Defensive {
		if resp.Header.Get(h) != "" {
			res.PresentHeaders = append(res.PresentHeaders, h)
		} else {
			res.MissingHeaders = append(res.MissingHeaders, h)
		}
	}
	// CanonicalHeaderKey: Get is case-insensitive, keep response spelling like Python.
	for _, h := range cfg.Leaky {
		if v := resp.Header.Get(h); v != "" {
			res.LeakyHeaders[h] = v
			if m := versionRE.FindStringSubmatch(v); m != nil {
				res.VersionDisclosures[h] = VersionDisclosure{Header: h, Software: m[1], Version: m[2]}
			}
		}
	}
	if res.MissingHeaders == nil {
		res.MissingHeaders = []string{}
	}
	if res.PresentHeaders == nil {
		res.PresentHeaders = []string{}
	}
	if chain == nil {
		chain = []string{}
	}
	finalURL := url
	if resp.Request != nil && resp.Request.URL != nil {
		finalURL = resp.Request.URL.String()
	}
	res.RedirectInfo = RedirectInfo{FinalURL: finalURL, Redirected: len(chain) > 0, Chain: chain}
	return res
}
