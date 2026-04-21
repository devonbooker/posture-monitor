package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func newServer(db *sql.DB) *http.Server {
	mux := http.NewServeMux()

	mux.Handle("/", http.FileServer(http.Dir("static")))
	mux.Handle("/metrics", promhttp.Handler())

	mux.HandleFunc("/api/results", func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(`
			SELECT url, up, status_code, latency_ms, checked_at
			FROM ping_results
			ORDER BY checked_at DESC
			LIMIT 100
		`)
		if err != nil {
			http.Error(w, "database error", http.StatusInternalServerError)
			log.Println("query error:", err)
			return
		}
		defer rows.Close()

		results := []PingResult{}
		for rows.Next() {
			var pr PingResult
			if err := rows.Scan(&pr.URL, &pr.Up, &pr.StatusCode, &pr.LatencyMs, &pr.CheckedAt); err != nil {
				log.Println("scan error:", err)
				continue
			}
			results = append(results, pr)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	})

	return &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}
