package main

import "github.com/prometheus/client_golang/prometheus"

var (
	pingTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "uptime_pings_total",
			Help: "Total number of pings per URL and status",
		},
		[]string{"url", "status"},
	)

	pingLatency = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "uptime_latency_ms",
			Help: "Most recent ping latency in milliseconds",
		},
		[]string{"url"},
	)

	tlsCertDaysUntilExpiry = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "tls_cert_days_until_expiry",
			Help: "Days until the leaf TLS certificate expires for each URL",
		},
		[]string{"url"},
	)

	securityHeaderPresent = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "security_header_present",
			Help: "1 if the given security response header is set, 0 if missing",
		},
		[]string{"url", "header"},
	)

	tlsInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "tls_info",
			Help: "Negotiated TLS protocol and cipher per URL. Value is always 1 - label values carry the signal.",
		},
		[]string{"url", "protocol", "cipher"},
	)

	tlsWeakProtocol = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "tls_weak_protocol",
			Help: "1 if the URL negotiated a protocol weaker than TLS 1.2, else 0",
		},
		[]string{"url"},
	)

	tlsWeakCipher = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "tls_weak_cipher",
			Help: "1 if the URL negotiated a cipher on Go stdlib's tls.InsecureCipherSuites() list, else 0",
		},
		[]string{"url"},
	)
)

func init() {
	prometheus.MustRegister(pingTotal)
	prometheus.MustRegister(pingLatency)
	prometheus.MustRegister(tlsCertDaysUntilExpiry)
	prometheus.MustRegister(securityHeaderPresent)
	prometheus.MustRegister(tlsInfo)
	prometheus.MustRegister(tlsWeakProtocol)
	prometheus.MustRegister(tlsWeakCipher)
}
