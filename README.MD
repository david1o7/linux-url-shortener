# Linux URL Shortener

A URL shortener built in Go while learning backend engineering, system design, and cloud engineering concepts.

The goal wasn't just to make a URL shortener. I used this project to learn how production backend services are structured by adding features one at a time instead of building everything at once.

## Features

- Create short URLs
- Redirect short URLs to their original destination
- PostgreSQL for persistent storage
- Redis caching for faster redirects
- Redis-based rate limiting
- Structured logging using slog
- Dockerized application
- Database indexing on the shortcode column
- Click analytics (tracks number of redirects)

## Tech Stack

- Go
- PostgreSQL
- Redis
- Docker
- Docker Compose

## Project Structure

```
.
├── cmd/
├── internal/
│   ├── cache/
│   ├── database/
│   ├── handlers/
│   ├── logger/
│   ├── middleware/
│   ├── models/
│   └── services/
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── go.sum
```

## Running the project

Clone the repository

```bash
git clone https://github.com/YOUR_USERNAME/linux-url-shortener.git

cd linux-url-shortener
```

Start everything

```bash
docker compose up --build
```

The application will be available at

```
http://localhost:8080
```

## API

### Create a short URL

```
POST /shorten
```

Example request

```json
{
  "url": "https://google.com"
}
```

Example response

```json
{
  "short_url": "http://localhost:8080/abc123"
}
```

### Redirect

```
GET /abc123
```

Redirects to the original URL.

## Why I built this

I'm currently learning Go, backend engineering, cloud engineering, and system design.

Instead of following tutorials forever, I'm building projects that introduce concepts like caching, rate limiting, Docker, indexing, and logging.

This project is part of that learning journey.

## Things I learned

- Building REST APIs in Go
- Working with PostgreSQL
- Redis caching
- Rate limiting
- Database indexing
- Docker and Docker Compose
- Structured logging
- Project organization
- Debugging backend applications

## Future Improvements

- User accounts
- URL expiration
- Custom aliases
- QR code generation
- Prometheus metrics
- Health check endpoint
- CI/CD with GitHub Actions
- Kubernetes deployment

## License

MIT