# Stage 1: build the Go binary
FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# CGO_ENABLED=0 produces a fully static binary - works on any linux base
# -trimpath strips the build machine's filesystem paths from the binary
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o uptime-monitor .

# Stage 2: minimal runtime image
FROM alpine:latest

# OCI labels - GHCR uses these to link the image back to the source repo
LABEL org.opencontainers.image.source="https://github.com/devonbooker/personal-uptime-monitor"
LABEL org.opencontainers.image.description="Self-hosted endpoint security monitor - uptime, TLS expiry, security headers"
LABEL org.opencontainers.image.licenses="MIT"

# ca-certificates lets the binary verify TLS certs when pinging HTTPS URLs
RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /app/uptime-monitor .
COPY static/ static/

EXPOSE 8080
CMD ["./uptime-monitor"]
