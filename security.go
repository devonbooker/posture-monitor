package main

import (
	"fmt"
	"net/http"
	"time"
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
