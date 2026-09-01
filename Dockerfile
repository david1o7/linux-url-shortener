# ---- Build stage ----
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache ca-certificates git

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Static binary, stripped debug info
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /out/server ./cmd

# ---- Runtime stage ----
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /

# Needed if the app runs migrations from disk at startup
COPY --from=builder /src/migrations /migrations
COPY --from=builder /out/server /server

ENV MIGRATIONS_DIR=/migrations
ENV APP_ENV=production

USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["/server"]