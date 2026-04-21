# Stage 1: build the Go binary
FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# CGO_ENABLED=0 produces a fully static binary - works on any linux base
# -trimpath strips the build machine's filesystem paths from the binary
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o posture-monitor .

# Stage 2: minimal runtime image
FROM alpine:latest

# OCI labels - GHCR uses these to link the image back to the source repo
LABEL org.opencontainers.image.source="https://github.com/devonbooker/posture-monitor"
LABEL org.opencontainers.image.description="Self-hosted endpoint security posture monitor - uptime, TLS expiry, security headers"
LABEL org.opencontainers.image.licenses="MIT"

# ca-certificates lets the binary verify TLS certs when pinging HTTPS URLs.
# nonroot user + group at fixed uid/gid 65532 lines up with the distroless
# convention and lets compose enforce a non-root container.
RUN apk add --no-cache ca-certificates && \
    addgroup -g 65532 -S nonroot && \
    adduser -u 65532 -S nonroot -G nonroot

WORKDIR /app
COPY --from=builder /app/posture-monitor .
COPY static/ static/

# /app/data is the SQLite volume mount point - create it in the image so its
# ownership is set before docker mounts the volume on top.
RUN mkdir -p /app/data && chown -R 65532:65532 /app

USER 65532:65532
EXPOSE 8080
CMD ["./posture-monitor"]
