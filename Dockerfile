# Stage 1: build the binary
FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o uptime-monitor .

# Stage 2: minimal runtime image
FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/uptime-monitor .
COPY static/ static/

EXPOSE 8080
CMD ["./uptime-monitor"]
