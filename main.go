package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"
)

func ping(db *sql.DB, url string) {
	start := time.Now()

	resp, err := http.Get(url)

	elapsed := time.Since(start)

	if err != nil {
		fmt.Printf("DOWN  %s — error: %v\n", url, err)
		saveResult(db, PingResult{URL: url, Up: false, CheckedAt: start})
		pingTotal.WithLabelValues(url, "down").Inc()
		return
	}
	defer resp.Body.Close()

	up := resp.StatusCode >= 200 && resp.StatusCode < 400
	result := PingResult{
		URL:        url,
		Up:         up,
		StatusCode: resp.StatusCode,
		LatencyMs:  elapsed.Milliseconds(),
		CheckedAt:  start,
	}
	saveResult(db, result)

	status := "up"
	if !up {
		status = "down"
	}
	pingTotal.WithLabelValues(url, status).Inc()
	pingLatency.WithLabelValues(url).Set(float64(elapsed.Milliseconds()))

	if up {
		fmt.Printf("UP    %s — status: %d, latency: %dms\n", url, resp.StatusCode, elapsed.Milliseconds())
	} else {
		fmt.Printf("DOWN  %s — status: %d, latency: %dms\n", url, resp.StatusCode, elapsed.Milliseconds())
	}
}

func main() {
	db := initDB()
	defer db.Close()

	urls := []string{
		"https://www.linkedin.com/feed/",
		"https://kodekloud.com",
		"https://startups.gallery",
		"https://www.nextplayjobs.com",
		"https://www.welcometothejungle.com/en/jobs",
		"https://wellfound.com",
	}

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	go startServer(db)

	fmt.Println("Starting uptime monitor — pinging every 30 seconds. Press Ctrl+C to stop.")

	// Ping immediately on startup, then again on each tick
	for _, url := range urls {
		ping(db, url)
	}

	for range ticker.C {
		fmt.Println("---")
		for _, url := range urls {
			go ping(db, url)
		}
	}
}
