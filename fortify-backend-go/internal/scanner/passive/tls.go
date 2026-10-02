package passive

import (
	"crypto/tls"
	"net"
	"time"
)

// TLSResult mirrors Python scanner/passive/tls.py output keys.
type TLSResult struct {
	TLSVersion  *string `json:"tls_version"`
	CertExpired *bool   `json:"cert_expired"`
	CertValid   bool    `json:"cert_valid"`
	CipherSuite *string `json:"cipher_suite"`
}

func strp(s string) *string { return &s }
func boolp(b bool) *bool    { return &b }

// ScanTLS opens a raw TLS connection to host:443 and reports the negotiated
// version, cipher, and certificate validity. Any failure (plain-HTTP target,
// closed port, timeout) yields the same null-shape the Python version
// returns so the dashboard renders identically.
func ScanTLS(hostname string) TLSResult {
	if hostname == "" {
		return TLSResult{CertValid: false}
	}
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	conn, err := tls.DialWithDialer(dialer, "tcp", net.JoinHostPort(hostname, "443"), &tls.Config{
		MinVersion: tls.VersionTLS10,
	})
	if err != nil {
		return TLSResult{CertValid: false}
	}
	defer conn.Close()

	state := conn.ConnectionState()
	version := tlsVersionName(state.Version)
	cipher := tls.CipherSuiteName(state.CipherSuite)

	var expired *bool
	valid := false
	if len(state.PeerCertificates) > 0 {
		exp := time.Now().After(state.PeerCertificates[0].NotAfter)
		expired = boolp(exp)
		valid = !exp
	}
	return TLSResult{
		TLSVersion:  strp(version),
		CertExpired: expired,
		CertValid:   valid,
		CipherSuite: strp(cipher),
	}
}

func tlsVersionName(v uint16) string {
	switch v {
	case tls.VersionTLS10:
		return "TLSv1"
	case tls.VersionTLS11:
		return "TLSv1.1"
	case tls.VersionTLS12:
		return "TLSv1.2"
	case tls.VersionTLS13:
		return "TLSv1.3"
	default:
		return "unknown"
	}
}
