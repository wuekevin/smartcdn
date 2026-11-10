# ── Build stage ──────────────────────────────────────────────────────────────
FROM golang:1.23-alpine AS builder

WORKDIR /src

# Cache dependencies separately from source changes.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the main server binary with all optimisations.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /out/smartcdn ./cmd/smartcdn

# Build the node-agent binary.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /out/node-agent ./cmd/node-agent

# Build the probe-agent binary.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /out/probe-agent ./cmd/probe-agent

# ── Runtime stage ─────────────────────────────────────────────────────────────
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /out/smartcdn /app/smartcdn
COPY --from=builder /src/migrations /app/migrations

# Non-root user
RUN addgroup -S smartcdn && adduser -S -G smartcdn smartcdn
USER smartcdn

EXPOSE 8080

ENTRYPOINT ["/app/smartcdn"]
