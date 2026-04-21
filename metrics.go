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
)

func init() {
	prometheus.MustRegister(pingTotal)
	prometheus.MustRegister(pingLatency)
	prometheus.MustRegister(tlsCertDaysUntilExpiry)
	prometheus.MustRegister(securityHeaderPresent)
}
