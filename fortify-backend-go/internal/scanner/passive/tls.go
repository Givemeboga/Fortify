package passive

import (
	"crypto/tls"
	"net"
	"strconv"
	"time"
)

// TLSResult mirrors the scanner output keys. VerificationError carries the
// handshake failure (expired, self-signed, wrong hostname, …) when the
// certificate was inspectable but untrusted.
type TLSResult struct {
	TLSVersion        *string `json:"tls_version"`
	CertExpired       *bool   `json:"cert_expired"`
	CertValid         bool    `json:"cert_valid"`
	CipherSuite       *string `json:"cipher_suite"`
	VerificationError *string `json:"verification_error,omitempty"`
}

func strp(s string) *string { return &s }
func boolp(b bool) *bool    { return &b }

// ScanTLS connects to host:port and reports the negotiated version, cipher,
// and certificate validity. A failed *verified* handshake no longer yields
// all-nulls: a second unverified handshake inspects the real certificate, so
// expired/self-signed/wrong-host certs are reported as findings instead of
// vanishing. Only a truly unreachable port returns the null shape.
func ScanTLS(hostname string, port int) TLSResult {
	if hostname == "" {
		return TLSResult{CertValid: false}
	}
	if port <= 0 {
		port = 443
	}
	target := net.JoinHostPort(hostname, strconv.Itoa(port))
	dialer := &net.Dialer{Timeout: 10 * time.Second}

	var verifyErr error
	if conn, err := tls.DialWithDialer(dialer, "tcp", target, &tls.Config{
		MinVersion: tls.VersionTLS10,
		ServerName: hostname,
	}); err == nil {
		defer conn.Close()
		return verifiedResult(conn)
	} else {
		verifyErr = err
	}

	// Untrusted but present: read the actual cert without verifying it.
	if conn, err := tls.DialWithDialer(dialer, "tcp", target, &tls.Config{ //nolint:gosec // intentional: inspection only, after verified dial failed
		MinVersion:         tls.VersionTLS10,
		InsecureSkipVerify: true,
	}); err == nil {
		defer conn.Close()
		res := verifiedResult(conn)
		res.CertValid = false
		msg := verifyErr.Error()
		res.VerificationError = &msg
		return res
	}

	return TLSResult{CertValid: false}
}

func verifiedResult(conn *tls.Conn) TLSResult {
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
