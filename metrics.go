package main

import "github.com/prometheus/client_golang/prometheus"

var (
	// Counts how many times each URL was pinged, labeled by result (up/down)
	pingTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "uptime_pings_total",
			Help: "Total number of pings per URL and status",
		},
		[]string{"url", "status"},
	)

	// Tracks the most recent latency for each URL
	pingLatency = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "uptime_latency_ms",
			Help: "Most recent ping latency in milliseconds",
		},
		[]string{"url"},
	)
)

func init() {
	prometheus.MustRegister(pingTotal)
	prometheus.MustRegister(pingLatency)
}
