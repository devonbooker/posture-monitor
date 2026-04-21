package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

func ping(ctx context.Context, db *sql.DB, url string) {
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		fmt.Printf("DOWN  %s - request build error: %v\n", url, err)
		saveResult(db, PingResult{URL: url, Up: false, CheckedAt: start})
		pingTotal.WithLabelValues(url, "down").Inc()
		return
	}

	resp, err := httpClient.Do(req)
	elapsed := time.Since(start)

	if err != nil {
		fmt.Printf("DOWN  %s - error: %v\n", url, err)
		saveResult(db, PingResult{URL: url, Up: false, CheckedAt: start})
		pingTotal.WithLabelValues(url, "down").Inc()
		return
	}
	defer resp.Body.Close()

	recordTLSExpiry(url, resp)
	recordSecurityHeaders(url, resp)

	up := resp.StatusCode >= 200 && resp.StatusCode < 400
	saveResult(db, PingResult{
		URL:        url,
		Up:         up,
		StatusCode: resp.StatusCode,
		LatencyMs:  elapsed.Milliseconds(),
		CheckedAt:  start,
	})

	status := "up"
	if !up {
		status = "down"
	}
	pingTotal.WithLabelValues(url, status).Inc()
	pingLatency.WithLabelValues(url).Set(float64(elapsed.Milliseconds()))

	label := "UP  "
	if !up {
		label = "DOWN"
	}
	fmt.Printf("%s  %s - status: %d, latency: %dms\n", label, url, resp.StatusCode, elapsed.Milliseconds())
}

func pingAll(ctx context.Context, wg *sync.WaitGroup, db *sql.DB, urls []string) {
	for _, url := range urls {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			ping(ctx, db, u)
		}(url)
	}
}

func pingInterval() time.Duration {
	if v := os.Getenv("PING_INTERVAL"); v != "" {
		if parsed, err := time.ParseDuration(v); err == nil && parsed > 0 {
			return parsed
		}
		fmt.Printf("WARN  ignoring invalid PING_INTERVAL=%q, using default\n", v)
	}
	return 30 * time.Second
}

func main() {
	db := initDB()
	defer db.Close()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	urls := []string{
		"https://www.linkedin.com/feed/",
		"https://kodekloud.com",
		"https://startups.gallery",
		"https://www.nextplayjobs.com",
		"https://www.welcometothejungle.com/en/jobs",
		"https://wellfound.com",
	}

	srv := newServer(db)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("server error: %v\n", err)
		}
	}()

	interval := pingInterval()
	fmt.Printf("uptime monitor started - interval=%s, targets=%d\n", interval, len(urls))

	var wg sync.WaitGroup
	pingAll(ctx, &wg, db, urls)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("shutdown signal received - waiting for in-flight pings...")
			wg.Wait()
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := srv.Shutdown(shutdownCtx); err != nil {
				fmt.Printf("server shutdown error: %v\n", err)
			}
			fmt.Println("clean exit.")
			return
		case <-ticker.C:
			fmt.Println("---")
			pingAll(ctx, &wg, db, urls)
		}
	}
}
