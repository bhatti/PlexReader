# Build stage — needs CGO for sqlite3 FTS5
FROM golang:1.23-alpine AS builder
RUN apk add --no-cache gcc musl-dev sqlite-dev

ARG VERSION=dev

WORKDIR /build
# Copy only backend source (context is repo root — see docker-compose.yaml)
COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ .
RUN CGO_ENABLED=1 go build -tags fts5 -ldflags="-s -w -X main.Version=${VERSION}" -o plexreader ./cmd/server

# Runtime stage — minimal image
FROM alpine:3.20
RUN apk add --no-cache ca-certificates sqlite-libs tzdata wget
RUN adduser -D -s /bin/sh plexreader

WORKDIR /app
COPY --from=builder /build/plexreader .

RUN mkdir -p /data && chown plexreader:plexreader /data
USER plexreader
VOLUME /data

EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=5s --retries=3 \
  CMD wget -qO- http://localhost:8080/healthz || exit 1

CMD ["./plexreader"]
