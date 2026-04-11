package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	_ "modernc.org/sqlite"
)

type PingResult struct {
	URL        string
	Up         bool
	StatusCode int
	LatencyMs  int64
	CheckedAt  time.Time
}

func initDB() *sql.DB {
	path := os.Getenv("DB_PATH")
	if path == "" {
		path = "uptime.db"
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		log.Fatal("failed to open database:", err)
	}

	// SQLite only supports one writer at a time — this prevents "database is locked" errors
	db.SetMaxOpenConns(1)

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS ping_results (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			url         TEXT NOT NULL,
			up          BOOLEAN NOT NULL,
			status_code INTEGER,
			latency_ms  INTEGER,
			checked_at  DATETIME NOT NULL
		)
	`)
	if err != nil {
		log.Fatal("failed to create table:", err)
	}

	return db
}

func saveResult(db *sql.DB, result PingResult) {
	_, err := db.Exec(
		`INSERT INTO ping_results (url, up, status_code, latency_ms, checked_at) VALUES (?, ?, ?, ?, ?)`,
		result.URL, result.Up, result.StatusCode, result.LatencyMs, result.CheckedAt,
	)
	if err != nil {
		log.Println("failed to save result:", err)
	}
}
