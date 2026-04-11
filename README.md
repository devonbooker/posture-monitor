# Personal Uptime Monitor

This was my first time making a web application. And it wasn't easy. But I realized that most things that need to be secured on the cloud are web application infrastructure. I did some research and found that Golang is a desirable language in cloud computing because of its lightweightness and efficiency. It's also a pretty easy programming language to understand. So I decided to build an app that would continuously ping URLs on a schedule, stores the results in a database, exposes them via a REST API, and displays them on a live dashboard - with full observability through Prometheus and Grafana, containerized with Docker, and deployed to AWS.

## What This Builds

A production-grade uptime monitor written in Go. No third-party monitoring services. No managed platforms. Every component built, containerized, and deployed from scratch.

- HTTP pinger that checks 6 URLs every 30 seconds using goroutines
- SQLite database that persists every check result
- REST API that exposes results as JSON
- Frontend dashboard that auto-refreshes every 30 seconds
- Prometheus metrics endpoint tracking ping counts and latency
- Grafana dashboards for visualizing uptime and latency over time
- Multi-stage Docker build and Docker Compose for local orchestration
- Deployed to AWS EC2 via Docker Hub

**Tools:** Go, SQLite, Prometheus, Grafana, Docker, Docker Compose, AWS EC2, Git, WSL2

---

## Prerequisites

- Go 1.21+
- Docker and Docker Compose
- AWS account with an EC2 instance (for deployment)
- Docker Hub account (for pushing images)

## How to Run Locally

```bash
# Clone the repo
git clone https://github.com/devonbooker/personal-uptime-monitor
cd personal-uptime-monitor

# Run the Go app directly
go run .

# Or run the full stack with Docker Compose
docker compose up
```

| Service | URL |
|---|---|
| Dashboard | `http://localhost:8080` |
| JSON API | `http://localhost:8080/api/results` |
| Prometheus metrics | `http://localhost:8080/metrics` |
| Grafana | `http://localhost:3000` |

## How to Deploy

```bash
# Build and push the image
docker build -t devonbooker/uptime-monitor:latest .
docker push devonbooker/uptime-monitor:latest

# SSH into EC2 and pull the latest image
ssh -i ~/.ssh/uptime-monitor-key.pem ec2-user@<your-ec2-ip>
docker-compose pull && docker-compose up -d
```

---

## Build Log

### Phase 1: Go Fundamentals + HTTP Pinger

Started by learning Go's module system, package structure, and error handling. Coming from no backend experience, the biggest mental shift was Go's explicit error handling - no try/catch, every function that can fail returns an error and you check it immediately.

The first real feature was a `ping()` function: make an HTTP GET, record the status code and latency, print whether the URL was UP or DOWN. That's it. But getting that working end-to-end was the foundation everything else built on.

### Phase 2: Scheduling + Goroutines

A one-shot ping isn't useful. Added a `time.Ticker` to run every 30 seconds and a goroutine per URL so all 6 are pinged concurrently instead of sequentially. Goroutines are one of Go's best features - they're cheap, they're easy to reason about, and they made the scheduling logic simple.

### Phase 3: SQLite + Persistence

Printing to the terminal disappears when the process exits. Added SQLite to persist every result - URL, status, latency, timestamp. Used parameterized queries (`?` placeholders) from the start to avoid SQL injection. Hit a `SQLITE_BUSY` error immediately because multiple goroutines were writing simultaneously - fixed with `db.SetMaxOpenConns(1)` to serialize writes through a single connection.

### Phase 4: REST API + Frontend

Exposed the database over HTTP using Go's standard `net/http` package - no framework. One route (`/api/results`) queries the last 100 results and returns them as JSON. The frontend is vanilla HTML and JavaScript: fetch the API, render a table, refresh every 30 seconds. No React, no build tooling. Go serves the static files directly with `http.FileServer`.

### Phase 5: Prometheus + Grafana

Added a `/metrics` endpoint using the Prometheus Go client library. Defined two metrics: a counter for total pings (labeled by URL and status) and a gauge for current latency. Prometheus scrapes the endpoint on a schedule; Grafana queries Prometheus and renders the charts. Ran all three as services in Docker Compose on a shared network so they can reach each other by service name.

Ran into two issues worth noting: port 9090 was already in use on my machine, so I removed Prometheus's host port mapping entirely - Grafana reaches it over the internal Docker network and doesn't need it exposed. Started services separately at first and got `lookup prometheus: no such host` because they ended up on different networks. Fixed by defining an explicit network in Compose and always running `docker compose up` to start everything together.

### Phase 6: Docker + AWS EC2

Wrote a multi-stage Dockerfile - the first stage compiles the Go binary, the second copies just the binary into a minimal Alpine image. No compiler, no source code in the final image.

Tried building on EC2 first and hit `compose build requires buildx 0.17.0 or later` - Amazon Linux's Docker package ships an outdated buildx plugin. Switched to the correct pattern: build locally, push to Docker Hub, pull on EC2. The instance never needs to build anything.

Opened three ports in the EC2 security group: 22 (SSH), 8080 (app), 3000 (Grafana). Everything else stays closed.

---

## Design Decisions

| Decision | Why |
|---|---|
| Go instead of Python/Node | Lightweight binary, built-in concurrency, and desirable in cloud engineering roles |
| SQLite instead of Postgres | No server to manage for a personal project - one file, zero configuration |
| `SetMaxOpenConns(1)` | SQLite has file-level write locking; this prevents `SQLITE_BUSY` errors from concurrent goroutines |
| Parameterized queries | Prevents SQL injection - never interpolate variables directly into SQL strings |
| Multi-stage Docker build | Keeps the production image small; no Go compiler or source code ships to EC2 |
| Build locally, pull on EC2 | EC2's Docker ships with an outdated buildx; this is the correct production pattern anyway |
| No host port for Prometheus | Services on the same Docker network reach each other by name - no need to expose internal services to the host |
