package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var securityHeaders = map[string]string{
	"hsts":                    "Strict-Transport-Security",
	"csp":                     "Content-Security-Policy",
	"x_frame_options":         "X-Frame-Options",
	"x_content_type_options":  "X-Content-Type-Options",
	"referrer_policy":         "Referrer-Policy",
	"permissions_policy":      "Permissions-Policy",
}

// recordTLSExpiry pulls the leaf cert off a completed HTTPS response and
// publishes days-until-expiry as a Prometheus gauge. Logs a warning when
// the cert has 30 or fewer days remaining - the industry tripwire for
// "renew now before customers see an outage."
func recordTLSExpiry(url string, resp *http.Response) {
	if resp.TLS == nil || len(resp.TLS.PeerCertificates) == 0 {
		return
	}
	leaf := resp.TLS.PeerCertificates[0]
	days := time.Until(leaf.NotAfter).Hours() / 24
	tlsCertDaysUntilExpiry.WithLabelValues(url).Set(days)

	if days <= 30 {
		fmt.Printf("WARN  %s - TLS cert expires in %.1f days (%s)\n", url, days, leaf.NotAfter.Format(time.RFC3339))
	}
}

// recordSecurityHeaders flags which defense-in-depth response headers are
// present. Missing HSTS/CSP on a production site is a real finding, not
// just pedantry - these are what keep session cookies and stored XSS in
// check at the browser layer.
func recordSecurityHeaders(url string, resp *http.Response) {
	for label, header := range securityHeaders {
		val := 0.0
		if resp.Header.Get(header) != "" {
			val = 1.0
		}
		securityHeaderPresent.WithLabelValues(url, label).Set(val)
	}
}

// recordTLSCipher publishes the negotiated TLS protocol version and cipher
// suite for each URL, and raises weak-configuration flags. Protocols below
// TLS 1.2 and ciphers on Go's tls.InsecureCipherSuites() list (RC4, 3DES,
// CBC-SHA1, export) are the things that show up in compliance audits and
// break modern CSP / cookie-security assumptions.
func recordTLSCipher(url string, resp *http.Response) {
	if resp.TLS == nil {
		return
	}

	protocol := tls.VersionName(resp.TLS.Version)
	cipher := tls.CipherSuiteName(resp.TLS.CipherSuite)

	// Clear any prior (url, oldProtocol, oldCipher) series so renegotiation
	// doesn't leave a stale 1-valued gauge hanging around for this URL.
	tlsInfo.DeletePartialMatch(prometheus.Labels{"url": url})
	tlsInfo.WithLabelValues(url, protocol, cipher).Set(1)

	weakProto := 0.0
	if resp.TLS.Version < tls.VersionTLS12 {
		weakProto = 1.0
		fmt.Printf("WARN  %s - weak TLS protocol: %s\n", url, protocol)
	}
	tlsWeakProtocol.WithLabelValues(url).Set(weakProto)

	weakCipher := 0.0
	for _, c := range tls.InsecureCipherSuites() {
		if c.ID == resp.TLS.CipherSuite {
			weakCipher = 1.0
			fmt.Printf("WARN  %s - insecure TLS cipher: %s\n", url, cipher)
			break
		}
	}
	tlsWeakCipher.WithLabelValues(url).Set(weakCipher)
}
