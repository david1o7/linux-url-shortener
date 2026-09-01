# Linux URL Shortener

A production-inspired URL shortener written in **Go**.

This is not a CRUD demo. It is a deliberate practice project for how backend services are configured, secured, tested, observed, and deployed.

**Stack:** Go · PostgreSQL · Redis · nginx · Prometheus · Grafana · Docker · GitHub Actions


## V1 DESIGN
[v1 system design](images/System-Design1.png)

## What it does



 `POST`  `/shorten` 
 `GET`  `/{code}`  Redirect to the original URL 
 `GET`  `/health/live`  Process liveness 
 `GET`  `/health/ready` Readiness (Postgres + Redis) 
 `GET`  `/metrics`  Prometheus metrics 


## Features

### Core
- Shorten long URLs and redirect via short codes
- **PostgreSQL** as the durable source of truth
- **Redis** for redirect caching and rate limiting
- Unique index on `shortcode`; lookups by code and original URL

### Security
- URL validation (scheme, host, reachability)
- SSRF protection (private / loopback / link-local rejection)
- DNS checks behind a testable resolver interface
- Token-bucket rate limiting (Redis-backed)
- Request / correlation IDs (`X-Request-ID`)

### Reliability
- Validated config at startup (fail fast on missing secrets)
- DB connection pooling and timeouts
- Versioned SQL migrations
- Graceful shutdown with bounded timeout
- nginx reverse proxy (ready for multi-instance upstream)

### Observability
- Structured logging (`slog` — text in dev, JSON in production)
- Prometheus metrics + alert rules
- Grafana-ready Compose profile
- Liveness vs readiness probes

### Quality
- Unit tests (handlers, config, middleware, services, validator)
- Integration tests (real Postgres + Redis, Testcontainers)
- CI: `gofmt`, `go vet`, race detector, golangci-lint

---

## Architecture

```text
Client
  → nginx (:80)
    → Go app (:8080, internal)
         → PostgreSQL   (source of truth)
         → Redis        (cache + token bucket)
```

Redis is an optimization. Cache miss falls back to the database. Uniqueness is never decided by the cache alone.

## Design decision: DB as source of truth (STAR)

**Situation**  
Checking “does this short code exist?” in application memory or Redis, then inserting, races under concurrency. Two requests can both pass the check and collide.

**Task**  
Make short-code uniqueness correct without distributed locks.

**Action**  
Use a PostgreSQL **UNIQUE** constraint on `shortcode`. Insert first. On unique violation (`23505`), map the driver error to a domain error and retry with a new code. The database is the authority; the app reacts.

**Result**  
No TOCTOU gap between check and write. Collisions are handled, not hoped away. Redis stays optional for speed, not for correctness.

> Note: error mapping must match the driver in use (`lib/pq` → `*pq.Error`, not a pgx type). Getting that wrong surfaces as raw SQL errors in tests and production.

---

## Hard lessons

| Struggle | Resolution |
|----------|------------|
| Fixed-window rate limits felt blunt under bursts | Token bucket with atomic consume in Redis |
| Everyone shared one rate-limit bucket behind a proxy | nginx forwards `X-Real-IP` / `X-Forwarded-For`; limiter reads client IP |
| Integration tests slow and flaky | Shared `TestMain` containers; `TRUNCATE` + Redis flush between tests |
| Async click increment vs test teardown | Cleanup wait so the DB is not closed mid-goroutine |
| Config defaults hiding missing secrets | Required env vars validated at startup |

---

## Project layout

```text
cmd/                  # process entrypoint
internal/
  config/             # env load + validation
  database/           # postgres, migrations, repository
  cache/              # redis + in-memory helpers
  handlers/           # HTTP handlers
  middleware/         # logging, metrics, rate limit, request ID
  services/           # short-code generation
  validator/          # URL + SSRF checks
  metric/ logger/
migrations/           # versioned *.up.sql
nginx/                # reverse proxy config
prometheus/           # scrape config + alert rules
tests/                # unit tests
tests/integration/    # //go:build integration
.github/workflows/    # CI
```

---

## Quick start

### Docker (recommended)

```bash
git clone https://github.com/YOUR_USERNAME/Linux-url-shortener.git
cd Linux-url-shortener
cp .env.example .env   # set DB_PASSWORD and BASE_URL

docker compose up --build -d
curl -s http://localhost/health/ready
```

Optional observability:

```bash
docker compose --profile obs up -d
# Prometheus :9090  ·  Grafana :3000
```

### Local binary

```bash
go mod download
# Postgres + Redis running; valid .env
go run ./cmd
```

---

## Configuration

| Variable | Notes |
|----------|--------|
| `DB_PASSWORD` | **Required** — no empty default in serious deploys |
| `BASE_URL` | Public URL embedded in short links (Codespaces URL, VPS IP, or domain) |
| `APP_ENV` | `production` → JSON logs |
| `RATE_LIMIT_CAPACITY` / `RATE_LIMIT_REFILL_PER_SEC` | Token bucket |
| `MIGRATIONS_DIR` | Defaults to `migrations` (image may use `/migrations`) |

Wrong or missing required config fails at **startup**, not on the first request.

---

## API examples

**Shorten**

```bash
curl -s -X POST http://localhost/shorten \
  -H 'Content-Type: application/json' \
  -d '{"url":"https://example.com"}'
```

**Redirect**

```bash
curl -sI http://localhost/{shortcode}
```

**Health**

```bash
curl -s http://localhost/health/live
curl -s http://localhost/health/ready
```

---

## Tests

```bash
# Unit tests
go test ./... -race -count=1

# Integration (Docker required)
go test -tags=integration ./tests/integration/ -v -count=1 -timeout 15m
```

CI runs format, vet, unit tests (with race), and golangci-lint on `main` and `v2`.

---

## Deploy

### Now — GitHub Codespaces

Codespaces is the temporary public demo path when a classic VPS is unavailable.

1. Open the repo in a Codespace  
2. `docker compose up --build -d`  
3. Forward port **80**, set visibility **Public**  
4. Set `BASE_URL` to the forwarded HTTPS URL and recreate the app container  
5. Exercise `/health/ready`, `/shorten`, and redirects  

**Limits:** the environment sleeps when idle; the URL is not permanent. Honest for a portfolio demo, not a forever short-link service.

### Later — VPS / always-on VM

Same Compose file:

1. Install Docker  
2. Clone repo, set `DB_PASSWORD` and `BASE_URL`  
3. Publish only **80** (and **443** later); keep Postgres/Redis internal  
4. Optional free hostname + Let’s Encrypt for a trusted HTTPS URL  

A React + Tailwind UI can live on Vercel; **redirects stay on the API host**.

---

## Version history

| Stage | Focus |
|-------|--------|
| **v1** | Postgres, Redis, metrics, validation, Docker |
| **v2** | Config validation, pooling, migrations, token bucket, request IDs, CI, broader tests, nginx-oriented deploy |

---

## Roadmap

- Custom aliases and URL expiration  
- Click analytics beyond a counter  
- Auth for privileged routes  
- Alertmanager notifications  
- Hardened `/metrics` exposure  
- Always-on host when infrastructure access allows  

---

## Why this project exists

To practice production-shaped backend engineering: boundaries (cache vs truth), middleware, security checks, observability, CI, integration testing, and reverse-proxy deploy — including the failures (driver error types, racey uniqueness, flaky containers) that tutorials often skip.

---

## License

MIT (or your choice)
