# Linux URL Shortener

A production-inspired URL shortener in **Go**. Built to practice real backend work: reliability, security, observability, and deployability — not just CRUD.

---

## Demo
![V1 system_design](images/System-Design1.png)

## What it does

| Endpoint | Purpose |
|----------|---------|
| `POST /shorten` | Create a short link |
| `GET /{code}` | Redirect to the original URL |
| `GET /health/live` | Liveness probe |
| `GET /health/ready` | Readiness (DB + Redis) |
| `GET /metrics` | Prometheus metrics |

**Stack:** Go · PostgreSQL · Redis · nginx · Prometheus · Grafana · Docker

---

## Features

- **PostgreSQL** as durable source of truth (unique short codes)
- **Redis** redirect cache + token-bucket rate limiting
- **URL validation** + SSRF protection (private / loopback / link-local rejected)
- **Request / correlation IDs** on every request
- **Structured logging** (text in dev, JSON in production)
- **Graceful shutdown**, connection pooling, versioned migrations
- **nginx** reverse proxy (ready for multi-instance load balancing)
- **CI:** `gofmt`, `go vet`, race tests, golangci-lint
- Unit tests + integration tests (Postgres/Redis via Testcontainers)

---

## Architecture (short)

```text
Client → nginx → Go app → PostgreSQL (truth)
                      ↘ Redis (cache + rate limit)
```

Redis is an optimization. If the cache is cold or empty, the app still serves redirects from the database.

---

## Design story: DB as source of truth (STAR)

**Situation:** Checking “does this short code already exist?” only in the app (or only in Redis) races under concurrency — two requests can generate the same code and both “pass.”

**Task:** Make uniqueness correct under concurrent load without a distributed lock.

**Action:** Rely on a **UNIQUE** constraint on `shortcode` in PostgreSQL. Insert first; on unique violation (`23505`), map to a domain error and retry with a new code. The database decides conflicts; the app reacts.

**Result:** No TOCTOU gap between “check” and “insert.” Cache stays optional. Collisions are handled, not hoped away.

---

## Other hard lessons

| Struggle | What we did |
|----------|-------------|
| Fixed-window rate limit felt crude under bursts | Token bucket in Redis (atomic Lua-style consume) |
| `lib/pq` errors vs `pgx` error types | Map `*pq.Error` / `23505` to `ErrShortCodeExist` for the driver we actually use |
| Integration tests slow / flaky containers | Shared `TestMain` containers + `TRUNCATE` / Redis flush between tests |
| Async `IncrementClicks` vs test teardown | Short cleanup wait so the DB isn’t closed mid-goroutine |
| Everyone shared one rate-limit bucket behind a proxy | nginx sets `X-Forwarded-For` / `X-Real-IP`; limiter reads real client IP |

---

## Version path

| Branch / stage | Focus |
|----------------|--------|
| **v1** | Postgres, Redis, metrics, validation, Docker |
| **v2** | Config validation, pooling, migrations, token bucket, CI, request IDs, tests, nginx-ready deploy |

---

## Run locally

```bash
git clone https://github.com/YOUR_USERNAME/Linux-url-shortener.git
cd Linux-url-shortener
cp .env.example .env   # set DB_PASSWORD at minimum

docker compose up --build
```

- App (via nginx): `http://localhost`
- Prometheus: `http://localhost:9090`
- Grafana: `http://localhost:3000`

Without Docker:

```bash
go mod download
go run ./cmd
```

Requires local Postgres + Redis and a valid `.env`.

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
curl -s http://localhost/health/ready
```

---

## Config that matters

| Variable | Role |
|----------|------|
| `DB_PASSWORD` | Required — no silent empty secret |
| `BASE_URL` | Public URL embedded in short links (nginx / VPS IP / domain) |
| `RATE_LIMIT_*` | Token-bucket capacity, refill, TTL |
| `APP_ENV` | `production` → JSON logs |

Invalid or missing required config fails **at startup**, not on the first request.

---

## Tests

```bash
# Unit (default)
go test ./... -race -count=1

# Integration (Docker required)
go test -tags=integration ./tests/integration/ -v -count=1 -timeout 15m
```

---

## Deploy sketch (VPS + nginx)

1. VPS with Docker  
2. `BASE_URL=http://YOUR_VPS_IP` (or HTTPS hostname later)  
3. Publish **only** port 80 (and 443 later); keep Postgres/Redis internal  
4. Optional free hostname + Let’s Encrypt for a browser-trusted URL  

Frontend (React + Tailwind on Vercel) can call the API; **redirects stay on the API host**.

---

## Roadmap

Custom aliases · URL expiry · click analytics · auth · Alertmanager notifications · hardened `/metrics`

---

## What this project was for

Learning production-shaped backend work: interfaces, repositories, middleware, caching boundaries, security checks, observability, CI, and shipping behind a reverse proxy — with mistakes fixed in public (driver error types, racey uniqueness, flaky tests) instead of hidden behind a tutorial happy path.
