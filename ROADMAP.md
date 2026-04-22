# Roadmap

Phased plan, prioritized for **Security Infrastructure / DevSecOps / Cloud Security** signal in interviews. Roughly one focused session per phase.

---

## Phase 1 - Grafana as code + alert rules ✅ SHIPPED 2026-04-21

**What:** Provision dashboards and alert rules via JSON + Grafana's sidecar provisioning. One dashboard (Uptime & TLS Posture) with panels for uptime % per URL, latency p50/p99, cert days-until-expiry (color threshold at 30 days), TLS protocol + cipher table, security-header matrix. Alert rules: cert < 30d, `tls_weak_protocol == 1`, `tls_weak_cipher == 1`, header regressions.

**Why:** seven meaningful metrics already exist and there's nothing visual to show for them. Dashboards turn "I made this" into "look at this live telemetry" in an interview. Dashboards-as-code beats hand-configured dashboards in any observability-minded interview.

**Deliverable:** `grafana/provisioning/dashboards/posture.json` + `grafana/provisioning/datasources/prometheus.yml` mounted into the Grafana container via compose. Dashboard reachable at `https://grafana.devonbooker.dev/d/posture`.

**Effort:** ~2 hours.

---

## Phase 2 - Container + CI security hardening ✅ SHIPPED 2026-04-21

**What:**
- **Dockerfile:** run as non-root (`USER 65532:65532`), drop all capabilities, `read_only: true` rootfs with a `tmpfs` for `/tmp` if needed, add `security_opt: [no-new-privileges:true]` in compose
- **CI:** add Trivy image scan (fail PR on `HIGH`+), add `govulncheck` on the Go module, SHA-pin every GitHub Action reference (tag pins can be hijacked via retag), enable Dependabot for Actions and Go modules
- **New file:** `.github/dependabot.yml`

**Why:** "Container security" and "infrastructure-level containment" are on Devon's skill stack. This phase converts every one of those words into something measurable that an interviewer can see. DevSecOps interviews love "how do you scan images?" - this answers it.

**Deliverable:** green Trivy scan in Actions, SHA-pinned workflow, Dependabot PRs rolling in, app container reports `runAsUser=65532` and read-only filesystem at runtime.

**Effort:** ~3 hours.

---

## Phase 3 - Terraform the infrastructure ✅ SHIPPED 2026-04-22

**What:** Hetzner Cloud provider (`hetznercloud/hcloud`) for the CAX11 server + SSH keys; Cloudflare provider for the A records; `cloud-init` for Docker install and UFW/fail2ban bootstrap. Secrets stay in GitHub Actions. Add a `terraform plan` PR check.

**Why:** "Terraform" is on Devon's skill stack. Right now the infra is a pet - if the VM dies, it gets rebuilt by hand. After Phase 3 it's cattle: `terraform apply` reproduces everything from zero. Single biggest jump in "real-world infra engineer" signal this project can make.

**Deliverable:** `terraform/` directory that reproduces VM + firewall + DNS from scratch. `terraform destroy && apply` round-trip works.

**Effort:** ~4 hours.

---

## Phase 4 - Auto-rollback on failed health check

**What:** When the workflow's health-check step fails, SSH back to the VM and redeploy the previous `:<sha>`. Track "last known good" via a file on the VM or a git note.

**Why:** current rollback is "SSH in, edit compose by hand, restart." Real prod systems auto-rollback on health failure. Closes the last manual step in the deploy loop - solid SRE-adjacent interview material.

**Effort:** ~2 hours.

---

## Phase 5+ - Lower-priority polish

- **Structured logging with `log/slog`** (replace `fmt.Printf`) + Loki sidecar for log aggregation. Good observability signal, medium ROI.
- **HTTP method probing** (PUT/DELETE on read-only endpoints → expect 405). On the original README roadmap but a weak security signal compared to TLS posture.
- **Secret rotation** - rotate GHCR PAT + SSH key on a cadence, automated.
- **Postgres migration** - don't until SQLite actually starts hurting. Currently it doesn't.

---

## Recommended order

1 → 2 → 3 → 4. Phases 5+ are polish.

If only three fit: 1 + 2 + 3. If only two: 2 + 3.
