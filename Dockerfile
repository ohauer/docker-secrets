# Build stage
FROM golang:1.25.6-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build static binary with stripped symbols
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -a -ldflags '-s -w -extldflags "-static"' \
    -o secrets-sync ./cmd/secrets-sync

# Final stage
FROM scratch

# Copy CA certificates for HTTPS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy binary
COPY --from=builder /build/secrets-sync /app/secrets-sync

# Create non-root user (nobody)
USER 65534:65534

# Set working directory
WORKDIR /app

# Expose health/metrics port
EXPOSE 8080

# Run as non-root
ENTRYPOINT ["/app/secrets-sync"]
