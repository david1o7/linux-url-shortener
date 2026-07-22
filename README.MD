# Linux URL Shortener
A production-inspired URL shortener built in Go.
This project started as a simple "shorten a URL" application, but my goal quickly changed. Instead of building another CRUD project, I wanted to learn how backend services are actually built and operated.
Along the way I ended up learning about databases, caching, testing, observability, graceful shutdown, dependency injection, Docker, security and a lot of debugging.
Tech Stack
- Go
- PostgreSQL
- Redis
- Docker
- Prometheus
- GitHub Actions (planned)
- Grafana (planned)

## The System Architecture 
![System Architecture](images/System-Design1.png)

## IMAGES 
![Image](images/image.png)
![gif](images/Code_XoEEY9mPtk.gif)

## Features
- Generate short URLs
- Redirect shortened URLs
- PostgreSQL persistence
- Redis caching
- Indexed database lookups
- Rate limiting
- URL validation
- SSRF protection
- DNS abstraction
- HTTP client abstraction
- Health check endpoint
- Prometheus metrics
- Request counters
- Request latency histogram
- Structured logging
- Graceful shutdown
- Unit tests
- Integration tests
- Environment variable configuration

<<<<<<< HEAD
Trade offs present:
- Single point of failure (1 golang http server)
- 0 back-up databases 
- Redis Can easily get full
- No loadbalancing 
- Rate limiting uses Fixed Window counter which is easy to bypass

Edge cases Handled:
- Cache Misses and Cache Hits
- Check for duplicated short codes
- Health Endpoint for checking dependencies  
Amongst others.
=======
## What I Learned
This project taught me much more than shortening URLs.
Some of the topics I learned while building it:
- Dependency Injection
- Interfaces in Go
- Repository-style database access
- Redis caching
- SQL indexing
- Docker
- Environment variables
- Prometheus metrics
- Histograms vs Counters
- Graceful shutdown
- Contexts
- HTTP middleware
- Unit testing
- Integration testing
- SSRF protection
- DNS resolution
- HTTP clients
- Production logging

## Biggest Challenges
This project definitely wasn't built in one sitting.
Docker
Docker was completely new to me.
At first, even building the container felt confusing. I spent quite a while understanding how images, containers and Dockerfiles actually worked before finally getting the application running inside a container.
Rate Limiting
The first implementation immediately rate-limited every request.
It turned out to be a bug in my implementation rather than Redis itself.
Debugging that taught me a lot about middleware execution order.
Prometheus
I initially couldn't understand why my metrics page was blank.
The issue wasn't Prometheus.
It was my middleware chain.
I had created the metrics middleware but wasn't actually using it.
Fixing that taught me how middleware wrapping works in Go.
DNS Lookup Tests
This was probably the most frustrating part of the project.
I refactored my URL validator to use dependency injection so I could mock DNS lookups.
Getting the tests to work correctly took much longer than expected.
At one point the "valid URL" test kept failing and I couldn't figure out why.
Eventually I discovered the validator itself wasn't broken.
The requests were timing out because of my internet connection.
Increasing the timeout fixed the issue.
That bug taught me that sometimes the code isn't wrong—the environment is.
URL Validation
I also learned that validating URLs is much harder than I originally thought.
It's not enough to check if a string looks like a URL.
I had to think about:
- localhost
- loopback addresses
- private networks
- unsupported protocols
- malformed URLs
- SSRF attacks
- Graceful Shutdown
Before this project I had never thought about how servers stop.
Now I understand why production services wait for active requests to finish before exiting and why operating systems send signals like SIGINT and SIGTERM.
## What I'm Most Proud Of
I'm proud that I didn't stop after getting the project to "work."
I kept improving it by adding things that aren't visible to users but matter in production:
- caching
- metrics
- health checks
- logging
- testing
- graceful shutdown
- security improvements
Those features taught me much more than simply building another CRUD application.
Current Limitations
>>>>>>> 9b063e8 (Add Tests, Add Prometheus Histogram, Add New System Arch. Diagram, ReadMe, Add UrlValidator)

## These are things I still want to improve.
- Docker Compose needs more polish.
- Grafana dashboard hasn't been set up yet.
- Custom aliases are not supported.
- URL expiration isn't implemented.
- Click analytics aren't implemented.
- Distributed rate limiting isn't supported.
- No authentication.
- No admin dashboard.

## Future Improvements
- Grafana dashboards
- Custom aliases
- URL expiration
- Click analytics
- User accounts
- JWT authentication
- API documentation
- Distributed deployment
- CI/CD pipeline
- Kubernetes deployment



# Running the Project

Clone the repository

```bash
git clone https://github.com/YOUR_USERNAME/Linux-url-shortener.git
cd Linux-url-shortener
```

Install dependencies

```bash
go mod download
```

Create your environment variables

```bash
cp .env.example .env
```

Run

```bash
go run ./cmd/server
```

or

```bash
docker compose up --build
```

---

# API

## Create Short URL

```
POST /shorten
```

Example request

```json
{
    "url":"https://google.com"
}
```

---

## Redirect

```
GET /{shortcode}
```

---

## Health Check

```
GET /health
```

---

## Prometheus Metrics

```
GET /metrics
```

---


<<<<<<< HEAD
The project is still a work in progress.

Things I'm currently improving:

- Docker Compose setup still needs refinement.
- Prometheus works, but Grafana dashboards haven't been configured yet.
- No automated tests yet.
- No CI/CD pipeline.
- No production deployment yet.
- No authentication or user accounts.
- No URL expiration or custom aliases.

---

# Future Improvements

- Grafana dashboards
- Graceful shutdown
- GitHub Actions
- Unit testing
- Production deployment
- Custom shortcodes
- URL expiration
- QR code generation
- CI/CD with GitHub Actions
- Kubernetes deployment

---

# SEO Keywords

For anyone discovering this project:

Go URL Shortener, Golang URL Shortener, Backend Engineering, Go REST API, PostgreSQL, Redis Cache, Docker, Docker Compose, Prometheus, System Design, Linux Development, Distributed Systems, Backend Portfolio Project, Rate Limiting, URL Shortener API.

---

# Final Thoughts

This project isn't meant to be perfect.

It's a snapshot of my backend engineering journey.

The goal wasn't just to build a URL shortener—it was to understand the technologies and design decisions behind modern backend systems.

I'm continuing to improve it as I learn more about cloud engineering, distributed systems, and system design.
=======
## If I Built It Again
I would:
- design interfaces earlier
- write tests sooner
- separate packages earlier
- keep configuration in one place
- spend more time designing middleware before coding
Final Thoughts
This project became much bigger than I expected.
I started by trying to build a URL shortener.
I ended up learning how backend services are designed, monitored, tested and operated.
There are still improvements to make, but this project marks a big step for
>>>>>>> 9b063e8 (Add Tests, Add Prometheus Histogram, Add New System Arch. Diagram, ReadMe, Add UrlValidator)
