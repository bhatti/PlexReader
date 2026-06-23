# PlexReader — Installation Guide

## Table of Contents

- [Docker Compose (recommended)](#docker-compose-recommended)
- [Environment Variables Reference](#environment-variables-reference)
- [Local Development Install](#local-development-install)
- [Upgrading](#upgrading)
- [Backup and Restore](#backup-and-restore)
- [Reverse Proxy / TLS](#reverse-proxy--tls)

---

## Docker Compose (recommended)

Docker Compose is the simplest way to run PlexReader in production. The stack consists of two containers:

- **backend** — Go server, serves the Connect/gRPC API on port 8080.
- **frontend** — nginx container, serves the Flutter web bundle and proxies `/plexreader.v1.*` requests to the backend.

### 1. Clone or download

```bash
git clone https://github.com/plexreader/plexreader.git
cd plexreader
```

### 2. Configure environment (optional)

Copy the example env file and edit as needed:

```bash
cp .env.example .env
```

The defaults work out of the box for a single-user local deployment. For a public server, at minimum enable auth and set a strong token:

```bash
PLEXREADER_AUTH_ENABLED=true
PLEXREADER_AUTH_TOKEN=change-me-to-a-long-random-string
```

### 3. Start the stack

```bash
make docker-up-d
```

Or without Make:

```bash
docker compose up -d   # Docker 20.10+
# docker-compose up -d  # older standalone docker-compose
```

Open **http://localhost:3000** in your browser.

### docker-compose.yaml walkthrough

```yaml
services:
  backend:
    build:
      context: .
      dockerfile: backend/Dockerfile
    ports:
      - "8080:8080"          # API port — do not expose publicly without auth + TLS
    environment:
      - PLEXREADER_PORT=8080
      - PLEXREADER_DB_PATH=/data/plexreader.db
      - PLEXREADER_REFRESH_INTERVAL=15m
      - PLEXREADER_AUTH_ENABLED=false
    volumes:
      - plexreader-data:/data   # Named volume — persists across restarts
    healthcheck:
      test: ["CMD", "wget", "-qO-", "http://localhost:8080/healthz"]
      interval: 10s
      timeout: 5s
      retries: 3
    restart: unless-stopped

  frontend:
    build:
      context: .
      dockerfile: ui/Dockerfile
    ports:
      - "3000:80"            # Web UI — point your browser (or reverse proxy) here
    depends_on:
      backend:
        condition: service_healthy   # Waits for /healthz to return 200
    restart: unless-stopped

volumes:
  plexreader-data:           # Docker-managed named volume at /var/lib/docker/volumes/
```

Key points:

- The **backend** container does not restart automatically if you change environment variables — run `make docker-down && make docker-up-d` after editing `.env`.
- The **frontend** container is stateless. It only serves static files and the nginx proxy config; all state lives in the backend volume.
- The health check polls `/healthz` (returns `200 ok`) every 10 seconds. The frontend waits for three consecutive successes before starting.
- The nginx proxy route is `/plexreader.v1.` (the Connect-JSON path prefix), not `/v1/`.

---

## Environment Variables Reference

All configuration is done via environment variables. There are no config files inside the container.

| Variable | Default | Description |
|----------|---------|-------------|
| `PLEXREADER_PORT` | `8080` | TCP port the backend HTTP/2 server listens on. |
| `PLEXREADER_DB_PATH` | `./data/plexreader.db` | Filesystem path to the SQLite database file. The directory is created automatically if it does not exist. |
| `PLEXREADER_REFRESH_INTERVAL` | `15m` | How often the scheduler wakes to refresh feeds. Accepts Go duration strings: `5m`, `1h`, `30s`. Individual feeds can override this with `refresh_interval_seconds`. |
| `PLEXREADER_AUTH_ENABLED` | `false` | Set to `true` to require a Bearer token on every API request. When `false`, the API is fully open. |
| `PLEXREADER_AUTH_TOKEN` | _(empty)_ | The static Bearer token clients must send in the `Authorization: Bearer <token>` header. Only used when `PLEXREADER_AUTH_ENABLED=true`. Use a long random string (at least 32 characters). |

### Generating a secure token

```bash
# Linux / macOS
openssl rand -hex 32
# Example output: a3f1c8e2b4d6f0a9c1e3b5d7f2a4c6e8b0d2f4a6c8e0b2d4f6a8c0e2b4d6f8a0
```

Set the output as `PLEXREADER_AUTH_TOKEN` in your `.env` file.

---

## Local Development Install

### Prerequisites

Install the following tools before proceeding:

```bash
# Go 1.25 or later (with CGO support)
go version   # must print go1.25 or higher

# C compiler for CGO (SQLite FTS5 requires cgo)
# macOS:  Xcode Command Line Tools
xcode-select --install
# Ubuntu: apt install gcc

# Flutter 3.x
flutter --version   # must print Flutter 3.x.x

# buf (Protocol Buffer toolchain)
# macOS:
brew install bufbuild/buf/buf
# Linux / other: https://buf.build/docs/installation
buf --version
```

### Step-by-step setup

**1. Clone the repository**

```bash
git clone https://github.com/plexreader/plexreader.git
cd plexreader
```

**2. Install Go dependencies**

```bash
make tidy
```

**3. Install Flutter dependencies**

```bash
make ui-deps
```

**4. Generate proto code** (only needed after editing `.proto` files)

```bash
make proto
```

**5. Start the backend**

```bash
make dev-backend
# Backend is now listening on http://localhost:8080
```

**6. Start the Flutter frontend** (in a second terminal)

```bash
make dev-ui
# Flutter opens Chrome on http://localhost:3001
```

`DEV_API_URL` defaults to `http://localhost:8080`. Override with `DEV_API_URL=http://otherhost:8080 make dev-ui`.

**7. Run tests**

```bash
# Backend (requires CGO + FTS5)
make backend-test

# Flutter analysis
make ui-analyze

# Full CI gate (backend tests + Flutter analysis)
make test
```

### Running in any browser (no Docker)

`make dev-ui` opens Chrome because Flutter's hot-reload injector requires it. The compiled output is plain HTML/JS/CSS and works in any modern browser.

```bash
# 1. Build the Flutter web bundle
make ui-build

# 2. Serve the build output
cd ui/build/web && python3 -m http.server 3001

# 3. Open in any browser
open http://localhost:3001       # macOS
xdg-open http://localhost:3001  # Linux
```

The backend (`make dev-backend`) must be running on port 8080.

### Building production assets locally

```bash
# Build the Go binary
make backend-build
# Output: backend/bin/plexreader

# Build the Flutter web bundle
make ui-build
# Output: ui/build/web/

# Build Docker images
make docker-build
```

---

## Upgrading

### Docker Compose upgrade

```bash
# Pull latest source, rebuild, restart
git pull
make docker-build
make docker-down
make docker-up-d
```

To reset the database entirely (e.g. after a schema change that AutoMigrate can't handle):

```bash
make docker-down-v   # stops containers AND removes the data volume
make docker-up-d
```

GORM `AutoMigrate` runs on every startup and applies any new columns or indexes automatically. Existing data is preserved. Destructive migrations (column renames, table drops) are never performed automatically — check the release notes before upgrading across major versions.

### Local binary upgrade

```bash
git pull
make tidy
make build   # backend-build + ui-build
```

---

## Backup and Restore

PlexReader's entire state is a single SQLite file.

**Default location in Docker:** `/data/plexreader.db` inside the `plexreader-data` named volume.

### Backup

```bash
# Copy the database file out of the running container
docker cp $(docker compose ps -q backend):/data/plexreader.db ./backup-$(date +%Y%m%d).db

# Or use the SQLite online backup command (safe while the server is running — WAL mode)
sqlite3 /path/to/plexreader.db ".backup /path/to/backup.db"
```

### Restore

```bash
# Stop the backend first to avoid write conflicts
docker compose stop backend

# Overwrite the database in the volume
docker cp ./backup-20240101.db $(docker compose ps -q backend):/data/plexreader.db

# Restart
docker compose start backend
```

### Automated backups

For production, schedule a nightly cron job to copy the database to object storage:

```bash
# Example: back up to an S3-compatible bucket with rclone
0 3 * * * docker exec $(docker compose -f /srv/plexreader/docker-compose.yaml ps -q backend) \
  sqlite3 /data/plexreader.db ".backup /tmp/backup.db" && \
  rclone copy /tmp/backup.db r2:my-bucket/plexreader/
```

---

## Reverse Proxy / TLS

For a public deployment, terminate TLS at a reverse proxy and forward to the Docker stack. The backend and frontend both speak plain HTTP/1.1 and HTTP/2 cleartext (h2c), so no TLS configuration is needed inside the containers.

### Caddy (recommended)

```caddyfile
reader.example.com {
    reverse_proxy localhost:3000
}
```

Caddy provisions a Let's Encrypt certificate automatically.

### nginx

```nginx
server {
    listen 443 ssl http2;
    server_name reader.example.com;

    ssl_certificate     /etc/letsencrypt/live/reader.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/reader.example.com/privkey.pem;

    location / {
        proxy_pass http://localhost:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

The frontend nginx container proxies `/plexreader.v1.*` calls to `backend:8080` and serves the Flutter SPA for everything else. Point your reverse proxy at port 3000 only.

---

## Chrome Extension

### Build

```bash
make chrome-ext
# Copies ui/build/web/* into ui/chrome_extension/
```

### Load in Chrome

1. Navigate to `chrome://extensions/`
2. Enable **Developer mode** (toggle, top right)
3. Click **Load unpacked**
4. Select the `ui/chrome_extension/` directory

The extension popup opens the full PlexReader UI. It also injects a content script that detects RSS/Atom feed links on any page.

> The extension requires Chrome or Chromium. For other browsers, use the web app.

### Pointing the extension at a remote server

By default the extension connects to `http://localhost:8080`. To use a remote server, rebuild with the correct API URL:

```bash
cd ui && ~/flutter/bin/flutter build web \
  --dart-define=API_BASE_URL=https://reader.example.com
cp -r build/web/* chrome_extension/
```

Then reload the unpacked extension in `chrome://extensions/`.
