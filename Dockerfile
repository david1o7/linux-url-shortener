FROM golang:1.25.0 AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/server ./cmd

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /
COPY --from=builder /src/migrations /migrations
COPY --from=builder /out/server /server
ENV MIGRATIONS_DIR=/migrations

ENV APP_ENV=production
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/server"]