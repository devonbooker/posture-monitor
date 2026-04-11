# Personal Uptime Monitor
### A first-timer's guide to web apps, databases, observability, and cloud deployment

---

## Overview

This document covers everything involved in building and deploying a personal uptime monitor from scratch. The app pings a list of URLs on a schedule, stores the results in a database, exposes them via a REST API, and displays them on a live dashboard — with full observability through Prometheus and Grafana, containerized with Docker, and deployed to AWS.

**Stack:**
- Go (Golang) — backend language
- SQLite — database
- Prometheus — metrics collection
- Grafana — metrics visualization
- Docker + Docker Compose — containerization
- AWS EC2 — cloud deployment

---

## Chapter 1: Setting Up a Go Project

### Go modules

Every Go project starts with a module, which is how Go tracks the project identity and its dependencies. You initialize one with:

```bash
go mod init uptime-monitor
```

This creates a `go.mod` file — similar to `package.json` in Node.js or `requirements.txt` in Python. As you add libraries, Go updates this file automatically.

### The entry point

Every Go program has a `main` package and a `main()` function. This is where execution begins:

```go
package main

func main() {
    // program starts here
}
```

### Key Go concepts introduced

**Imports** — Go uses explicit imports from the standard library or third-party packages:
```go
import (
    "fmt"
    "net/http"
    "time"
)
```

**Functions** — defined with `func`, with typed parameters and return values:
```go
func ping(url string) {
    // ...
}
```

**Error handling** — Go doesn't use try/catch. Functions return errors explicitly and you check them manually:
```go
resp, err := http.Get(url)
if err != nil {
    // handle the error
}
```

**`defer`** — runs a statement when the surrounding function exits. Used to clean up resources:
```go
defer resp.Body.Close()
```

---

## Chapter 2: Making HTTP Requests

The first real feature was a `ping()` function that makes an HTTP GET request and reports whether a URL is up or down.

```go
func ping(url string) {
    start := time.Now()
    resp, err := http.Get(url)
    elapsed := time.Since(start)

    if err != nil {
        fmt.Printf("DOWN  %s — error: %v\n", url, err)
        return
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 200 && resp.StatusCode < 400 {
        fmt.Printf("UP    %s — status: %d, latency: %dms\n", url, resp.StatusCode, elapsed.Milliseconds())
    } else {
        fmt.Printf("DOWN  %s — status: %d, latency: %dms\n", url, resp.StatusCode, elapsed.Milliseconds())
    }
}
```

### What you learned
- `http.Get()` sends an HTTP GET request and returns a response and an error
- HTTP status codes 200–399 indicate success; anything else is a problem
- `time.Now()` and `time.Since()` measure how long something takes (latency)
- A status code of 200 means "OK", 404 means "not found", 500 means "server error"

---

## Chapter 3: Scheduling with Tickers and Goroutines

Running the ping once and exiting isn't useful. The app needed to ping on a repeating schedule.

### Tickers

A `time.Ticker` fires on a fixed interval. `ticker.C` is a **channel** — a pipe that delivers a value every time the ticker fires. `for range ticker.C` blocks and waits, running the loop body on each tick:

```go
ticker := time.NewTicker(30 * time.Second)
defer ticker.Stop()

for range ticker.C {
    for _, url := range urls {
        ping(url)
    }
}
```

### Goroutines

The `go` keyword launches a **goroutine** — Go's lightweight thread. Without it, each URL would wait for the previous one to finish before starting. With it, all URLs are pinged simultaneously:

```go
for _, url := range urls {
    go ping(url)
}
```

This is one of Go's greatest strengths. Goroutines are cheap (you can run thousands of them) and they make concurrent programming straightforward.

---

## Chapter 4: Storing Results in SQLite

Printing to the terminal isn't enough — results need to persist. SQLite is a file-based database that requires no server to run, making it ideal for a project like this.

### The database driver

Go's `database/sql` package provides a standard interface for databases. The `modernc.org/sqlite` driver is pure Go and requires no extra installation:

```bash
go get modernc.org/sqlite
```

The `_` import is used to register the driver without calling it directly:
```go
import _ "modernc.org/sqlite"
```

### Structs

A **struct** is Go's way of grouping related data — like a row in a database table:

```go
type PingResult struct {
    URL        string
    Up         bool
    StatusCode int
    LatencyMs  int64
    CheckedAt  time.Time
}
```

### Initializing the database

```go
func initDB() *sql.DB {
    db, err := sql.Open("sqlite", "uptime.db")
    // ...
    db.SetMaxOpenConns(1) // important — see troubleshooting below
    db.Exec(`CREATE TABLE IF NOT EXISTS ping_results (...)`)
    return db
}
```

### Inserting rows

```go
db.Exec(
    `INSERT INTO ping_results (url, up, status_code, latency_ms, checked_at) VALUES (?, ?, ?, ?, ?)`,
    result.URL, result.Up, result.StatusCode, result.LatencyMs, result.CheckedAt,
)
```

The `?` placeholders prevent SQL injection — never interpolate user input directly into SQL strings.

### Splitting code into files

At this point the code was split into two files:
- `main.go` — scheduling and ping logic
- `db.go` — database initialization and queries

Go compiles all files in a package together. You must run `go run .` (not `go run main.go`) to include all files.

---

## Chapter 5: Building a REST API

A REST API exposes your data over HTTP so other programs (like a browser) can fetch it.

### Registering routes

`http.HandleFunc` maps a URL path to a handler function:

```go
http.HandleFunc("/api/results", func(w http.ResponseWriter, r *http.Request) {
    // w = what you write the response into
    // r = the incoming request
})
```

### Querying and returning JSON

```go
rows, _ := db.Query(`SELECT ... FROM ping_results ORDER BY checked_at DESC LIMIT 100`)
defer rows.Close()

results := []PingResult{}
for rows.Next() {
    var r PingResult
    rows.Scan(&r.URL, &r.Up, &r.StatusCode, &r.LatencyMs, &r.CheckedAt)
    results = append(results, r)
}

w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(results)
```

`rows.Scan()` reads one database row into your struct fields, column by column. `json.NewEncoder(w).Encode(results)` converts your Go slice into JSON and writes it to the response.

### Starting the server

The server runs as a goroutine so it doesn't block the ticker loop:

```go
go startServer(db)
```

---

## Chapter 6: The Frontend

A simple HTML page with vanilla JavaScript fetches from the API and renders a table. No frameworks needed.

```javascript
async function loadResults() {
    const resp = await fetch('/api/results');
    const data = await resp.json();

    for (const row of data) {
        // build a table row for each result
    }
}

loadResults();
setInterval(loadResults, 30000); // refresh every 30 seconds
```

Go serves the static files with one line:

```go
http.Handle("/", http.FileServer(http.Dir("static")))
```

This tells Go to serve any file in the `static/` folder as-is. When the browser requests `/`, Go returns `static/index.html`.

---

## Chapter 7: Observability with Prometheus

Prometheus is a time-series database that scrapes metrics from your app on a schedule. Your app exposes a `/metrics` endpoint; Prometheus polls it.

### Installing the client library

```bash
go get github.com/prometheus/client_golang/prometheus
go get github.com/prometheus/client_golang/prometheus/promhttp
```

### Defining metrics

```go
var (
    pingTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{Name: "uptime_pings_total"},
        []string{"url", "status"},
    )
    pingLatency = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{Name: "uptime_latency_ms"},
        []string{"url"},
    )
)
```

**Counter** — only goes up. Good for "how many times did X happen."
**Gauge** — can go up or down. Good for "what is the current value of X."
**Labels** — dimensions you can filter by (e.g. filter `uptime_pings_total` by `url` or `status`).

### Recording metrics

```go
pingTotal.WithLabelValues(url, "up").Inc()
pingLatency.WithLabelValues(url).Set(float64(elapsed.Milliseconds()))
```

### Exposing the endpoint

```go
http.Handle("/metrics", promhttp.Handler())
```

---

## Chapter 8: Containerizing with Docker

Docker packages your app and its dependencies into an **image** that runs the same way everywhere.

### Multi-stage Dockerfile

```dockerfile
# Stage 1: compile the Go binary
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o uptime-monitor .

# Stage 2: copy just the binary into a tiny image
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/uptime-monitor .
COPY static/ static/
EXPOSE 8080
CMD ["./uptime-monitor"]
```

The two-stage build keeps the final image small — no Go compiler or source code in production.

### Docker Compose

Docker Compose lets you define and run multiple containers together. Each container is a **service**. Services on the same network can reach each other by service name:

```yaml
services:
  app:
    image: devonbooker/uptime-monitor:latest
    ports:
      - "8080:8080"
    networks:
      - monitor-net

  prometheus:
    image: prom/prometheus:latest
    networks:
      - monitor-net

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    networks:
      - monitor-net

networks:
  monitor-net:
```

Because all three services share `monitor-net`, Grafana can reach Prometheus at `http://prometheus:9090` — no IP address needed.

### Volumes

Named volumes persist data across container restarts:

```yaml
volumes:
  sqlite-data:
  prometheus-data:
  grafana-data:
```

Without volumes, your database would be wiped every time a container restarted.

### Environment variables

The database path is passed via environment variable so it works both locally and in Docker:

```go
path := os.Getenv("DB_PATH")
if path == "" {
    path = "uptime.db"
}
```

---

## Chapter 9: Deploying to AWS EC2

EC2 (Elastic Compute Cloud) is a virtual machine you rent from AWS. You SSH into it and run your app just like a local machine.

### The deployment workflow

1. Build the Docker image locally
2. Push it to Docker Hub (a public image registry)
3. SSH into EC2
4. Pull the image and run it with Docker Compose

```bash
# Local
docker build -t devonbooker/uptime-monitor:latest .
docker push devonbooker/uptime-monitor:latest

# On EC2
docker-compose pull
docker-compose up -d
```

The `-d` flag runs containers in the background (detached mode).

### Security groups

A security group is a firewall for your EC2 instance. You opened three ports:
- **22** — SSH (to connect to the instance)
- **8080** — your app
- **3000** — Grafana

### Key pairs

An SSH key pair authenticates you to EC2 without a password. The private key lives on your machine (`~/.ssh/uptime-monitor-key.pem`); AWS holds the public key. The `chmod 400` command restricts the file so only you can read it — SSH refuses to use keys with loose permissions.

---

## Troubleshooting Log

### `go run main.go` — undefined: saveResult, PingResult, initDB

**Cause:** When you specify a single file, Go only compiles that file. Functions defined in `db.go` are invisible.

**Fix:** Run `go run .` to compile all files in the directory.

---

### `database is locked (SQLITE_BUSY)`

**Cause:** Multiple goroutines tried to write to SQLite simultaneously. SQLite uses a file-level lock — only one write can happen at a time.

**Fix:** `db.SetMaxOpenConns(1)` forces Go's connection pool to queue writes through a single connection.

---

### Port 9090 refused to connect / `ports are not available`

**Cause:** Port 9090 was already in use on the host machine. Docker couldn't bind to it.

**Fix:** Removed the Prometheus port mapping entirely from `docker-compose.yml`. Grafana reaches Prometheus over the internal Docker network — it doesn't need to be exposed to the host.

---

### Grafana: `lookup prometheus: no such host`

**Cause:** Services were started separately (`docker compose up prometheus`, then others), so they ended up on different networks.

**Fix:** Added an explicit `monitor-net` network to `docker-compose.yml` and assigned all services to it. Always use `docker compose up` to start all services together.

---

### EC2: `compose build requires buildx 0.17.0 or later`

**Cause:** The version of Docker installed via Amazon Linux's package manager ships with an outdated buildx plugin.

**Fix:** Switched the deployment workflow to build locally and push to Docker Hub. The EC2 instance pulls the pre-built image instead of building from source. This is the correct pattern for production deployments.

---

## How to Improve It Further

### Short term

- **Add all 6 URLs from the start** — the monitor only tracked 3 initially
- **Configurable ping interval** — read it from an environment variable instead of hardcoding 30 seconds
- **HTTP timeouts** — `http.Get()` will wait forever if a server is slow. Set a timeout:
  ```go
  client := &http.Client{Timeout: 10 * time.Second}
  resp, err := client.Get(url)
  ```
- **Graceful shutdown** — handle `SIGINT` (Ctrl+C) so the app closes the DB cleanly before exiting

### Medium term

- **Alerting** — send an email or Slack message when a URL goes down. Grafana supports alert rules natively.
- **Grafana provisioning** — save your Grafana dashboards and data source config as code so they survive container restarts automatically
- **Uptime percentage** — calculate and display what % of checks returned UP over the last 24 hours
- **Postgres instead of SQLite** — SQLite works fine for a personal project, but Postgres handles concurrent writes better and is easier to scale. AWS RDS can host it.
- **HTTPS** — add a domain name and a TLS certificate (Let's Encrypt is free) so the app is accessible at `https://yourname.com` instead of a raw IP

### Long term

- **ECS instead of EC2** — AWS Elastic Container Service manages containers for you. No SSH, no manual updates — just push a new image and it deploys automatically.
- **CI/CD pipeline** — use GitHub Actions to automatically build, push, and deploy on every git push to main
- **Multi-region monitoring** — run the pinger from multiple AWS regions to detect regional outages
- **Historical charts** — add a Grafana time series panel showing uptime % over the past 30 days

---

## Quick Reference

### Local development

```bash
go run .                  # run the app
go build ./...            # verify it compiles
```

### Docker

```bash
docker compose up         # start all services
docker compose down       # stop all services
docker compose ps         # check what's running
docker compose logs app   # view logs for a service
```

### EC2

```bash
# SSH in
ssh -i ~/.ssh/uptime-monitor-key.pem ec2-user@54.205.243.228

# Deploy a new version
docker-compose pull && docker-compose up -d

# Stop the instance (saves money)
aws ec2 stop-instances --instance-ids i-02878ab3a62e5700a

# Start it again
aws ec2 start-instances --instance-ids i-02878ab3a62e5700a
```

### Endpoints

| URL | What it does |
|---|---|
| `http://54.205.243.228:8080` | Frontend dashboard |
| `http://54.205.243.228:8080/api/results` | Raw JSON results |
| `http://54.205.243.228:8080/metrics` | Prometheus metrics |
| `http://54.205.243.228:3000` | Grafana |
