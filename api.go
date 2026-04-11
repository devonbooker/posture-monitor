package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func startServer(db *sql.DB) {
	http.Handle("/", http.FileServer(http.Dir("static")))
	http.Handle("/metrics", promhttp.Handler())

	http.HandleFunc("/api/results", func(w http.ResponseWriter, r *http.Request) {
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
			var r PingResult
			err := rows.Scan(&r.URL, &r.Up, &r.StatusCode, &r.LatencyMs, &r.CheckedAt)
			if err != nil {
				log.Println("scan error:", err)
				continue
			}
			results = append(results, r)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	})

	log.Println("API server listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
