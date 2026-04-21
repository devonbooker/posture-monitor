# Endpoint Security Monitor

A self-hosted Go service that pings a set of target URLs on a schedule and records three things every security engineer cares about: is it up, how fresh is the TLS certificate, and which defense-in-depth response headers is it actually setting.

Built to run on a €4/mo Hetzner ARM64 VM with the full observability stack alongside it: Prometheus for metrics, Grafana for visualization, Docker Compose for orchestration, GHCR for the image pipeline.

---

## What It Monitors

For each target URL, every 30 seconds:

- **Availability** - HTTP status, latency, up/down result, persisted to SQLite
- **TLS certificate health** - parses the leaf cert off the live response, emits `tls_cert_days_until_expiry` as a Prometheus gauge, logs a warning at 30 days
- **TLS protocol and cipher posture** - reads the negotiated protocol version and cipher off `resp.TLS`, emits `tls_info{protocol, cipher}`, and raises `tls_weak_protocol` / `tls_weak_cipher` flags when a target negotiates TLS 1.0/1.1 or a cipher on Go stdlib's `tls.InsecureCipherSuites()` list (RC4, 3DES, CBC-SHA1, export)
- **Security header posture** - checks for HSTS, CSP, X-Frame-Options, X-Content-Type-Options, Referrer-Policy, Permissions-Policy and emits `security_header_present{url, header}` so you can alert on regressions

Why this matters: the #1 cause of public-facing outages isn't the application failing, it's the TLS cert expiring. The #1 cause of "how did they get stored XSS past us" is a missing or weak CSP. This tool puts both on the same dashboard as uptime.

---

## Stack

| Layer | Choice | Why |
|---|---|---|
| Language | Go 1.25 | Small static binary, goroutines for concurrent pings, strong stdlib for HTTP/TLS/crypto |
| Storage | SQLite (modernc.org/sqlite) | One file, zero config, pure Go driver (no CGO) |
| Metrics | Prometheus client_golang | Industry standard, scrapes `/metrics` |
| Dashboards | Grafana | PromQL-backed panels for uptime, latency, cert expiry |
| Runtime | Docker Compose | App + Prometheus + Grafana on one shared network |
| Image registry | GitHub Container Registry (ghcr.io) | Multi-arch (amd64 + arm64), ties image to repo |
| Host | Hetzner Cloud CAX11 (Ampere ARM64, Falkenstein) | €3.29/mo, 4GB RAM, plenty for the whole stack |

---

## Hardening Applied to the Code

- `http.Client{Timeout: 10s}` instead of `http.Get` - prevents goroutine leaks when a target stops responding mid-connection
- `http.NewRequestWithContext` - in-flight pings cancel on shutdown signal instead of leaking
- `signal.NotifyContext(SIGINT, SIGTERM)` + `WaitGroup` + `srv.Shutdown` - clean exit, no half-written DB rows
- `db.SetMaxOpenConns(1)` - SQLite has file-level write locking, this serializes concurrent goroutine writes through one connection
- Parameterized queries everywhere - no string-concat SQL
- `http.Server{ReadTimeout, WriteTimeout, IdleTimeout}` on the API - slowloris resistance
- Multi-stage Docker build - production image ships the binary and the static frontend, no Go toolchain, no source

## Hardening Applied to the VM

- Non-root user, key-only SSH, `PermitRootLogin no`, `PasswordAuthentication no`
- UFW default-deny inbound, only 22/8080/3000 open
- `fail2ban` enabled on sshd
- Prometheus port 9090 never exposed to the internet - reached over the internal Docker network by service name

---

## Configuration

| Env var | Default | Purpose |
|---|---|---|
| `DB_PATH` | `uptime.db` | SQLite file location. Override to a mounted volume in Docker. |
| `PING_INTERVAL` | `30s` | Go duration string (`10s`, `1m`, `5m`). Invalid values fall back to default with a warning. |
| `GRAFANA_ADMIN_PASSWORD` | (required) | Grafana admin password. Read from `.env` by `docker compose`. Compose will refuse to start if unset. |

Create a `.env` from the template before bringing the stack up:

```bash
cp .env.example .env
# edit .env and set a real password (openssl rand -base64 32)
```

---

## Run Locally

```bash
go run .
```

Or the full stack:

```bash
docker compose up
```

| Endpoint | What |
|---|---|
| `http://localhost:8080` | Dashboard |
| `http://localhost:8080/api/results` | Last 100 checks as JSON |
| `http://localhost:8080/metrics` | Prometheus scrape target |
| `http://localhost:3000` | Grafana (admin / admin on first login) |

---

## Deploy to Hetzner

Automated via GitHub Actions (`.github/workflows/deploy.yml`): every push to `main` builds a multi-arch image, pushes `:latest` and `:<short-sha>` tags to GHCR, SSHes to the VM, pulls, restarts, and health-checks the dashboard. Rollback is `docker compose` with the image tag pinned to a previous SHA.

### One-time VM setup

```bash
ssh devon@<vm-ip>

# Log in to GHCR once - creds cache in ~/.docker/config.json
echo $GHCR_PAT | docker login ghcr.io -u devonbooker --password-stdin

# Place compose files + .env on the VM
mkdir -p ~/posture-monitor && cd ~/posture-monitor
# scp docker-compose.yml prometheus.yml .env.example from your laptop into this dir
cp .env.example .env
# edit .env and set GRAFANA_ADMIN_PASSWORD to a real value

# Open ports for the dashboard and Grafana
sudo ufw allow 8080/tcp
sudo ufw allow 3000/tcp
```

### One-time GitHub Actions setup

In Settings → Secrets and variables → Actions, add three repo secrets:

| Secret | Value |
|---|---|
| `HETZNER_HOST` | VM public IP |
| `HETZNER_USER` | SSH user (e.g. `devon`) |
| `HETZNER_SSH_KEY` | Private key (full file contents, including `-----BEGIN`/`-----END`). Generate with `ssh-keygen -t ed25519 -f ~/.ssh/gh_actions_hetzner` and install the `.pub` on the VM via `ssh-copy-id`. |

Prometheus intentionally has no host port - Grafana reaches it over the internal Docker network.

---

## Metrics Reference

| Metric | Type | Labels | What it tells you |
|---|---|---|---|
| `uptime_pings_total` | Counter | `url`, `status` | How many checks per URL, split by up/down |
| `uptime_latency_ms` | Gauge | `url` | Most recent response time |
| `tls_cert_days_until_expiry` | Gauge | `url` | Days remaining on the leaf cert. Alert on `< 30`. |
| `tls_info` | Gauge | `url`, `protocol`, `cipher` | Always 1. Label values carry the negotiated TLS version (e.g. `TLS 1.3`) and cipher suite (e.g. `TLS_AES_128_GCM_SHA256`). |
| `tls_weak_protocol` | Gauge | `url` | 1 if URL negotiated a protocol weaker than TLS 1.2, else 0. |
| `tls_weak_cipher` | Gauge | `url` | 1 if URL negotiated a cipher on Go stdlib's insecure list, else 0. |
| `security_header_present` | Gauge | `url`, `header` | 1 if header set, 0 if missing. Alert on regression. |

Example PromQL for a "certs expiring soon" alert:

```promql
tls_cert_days_until_expiry < 30
```

Example for "weak TLS protocol" alert (TLS 1.0 or 1.1 negotiated):

```promql
tls_weak_protocol == 1
```

Example for "insecure cipher suite" alert (RC4, 3DES, CBC-SHA1, export):

```promql
tls_weak_cipher == 1
```

Example for "CSP regression" alert:

```promql
security_header_present{header="csp"} == 0
```

---

## Roadmap

- Grafana dashboards provisioned as code (currently configured by hand)
- Migrate SQLite to Postgres once retention exceeds what a single-writer file can handle
- Add HTTP method probing (verify `PUT`/`DELETE` return 405 on read-only endpoints)
- Grafana alert rules that page on cert expiry and security header regressions
