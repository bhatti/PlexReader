# PlexReader

**Self-hosted RSS reader inspired by Feedly**

[![License: LGPL-2.1](https://img.shields.io/badge/License-LGPL--2.1-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://go.dev)
[![Flutter](https://img.shields.io/badge/Flutter-3.44+-54C5F8?logo=flutter)](https://flutter.dev)

PlexReader is a self-hosted, open-source RSS/Atom feed reader. It stores everything in a single SQLite file, exposes a Connect/gRPC+REST API, and ships a Flutter web frontend with Magazine, Title, and Cards views.

---

## Features

- **Today feed** — articles from the last 24 hours across all subscriptions
- **Multiple view modes** — Magazine (image + summary), Title (compact list), Cards (grid)
- **OPML import/export** — migrate from Feedly, Inoreader, or any OPML-compatible reader
- **Full-text search** — SQLite FTS5 with Porter stemming and BM25 relevance ranking
- **Unread counts** — per-feed and per-folder badges
- **Keyboard shortcuts** — `j`/`k` navigate, `m` toggle read, `s` star, `l` save, `Shift+A` mark all read
- **Starred & Saved for Later** — permanent bookmarks and reading list
- **Recently Read** — history of articles you've opened
- **Conditional HTTP fetching** — ETag/Last-Modified to minimise bandwidth
- **SSRF protection** — DNS-rebinding safe; blocks private IP ranges and cloud metadata endpoints
- **HTML sanitisation** — bluemonday strips XSS from feed content before storage
- **Retention policy** — auto-prune old read articles; starred articles are never deleted
- **Optional auth** — Bearer token authentication (disabled by default)
- **Swagger UI** — browse and try the REST API at `/swagger/`
- **Dark theme** — Material 3 design with configurable preferences

---

## Quick Start — Docker (recommended)

```bash
git clone https://github.com/plexreader/plexreader.git
cd plexreader
make docker-up-d
```

Open **http://localhost:3000** — UI is up.  
Backend API and health check: **http://localhost:8080/healthz**

To stop: `make docker-down`

---

## Quick Start — Local Dev (no Docker)

**Prerequisites:** Go 1.23+, CGO (Xcode CLT on macOS / `gcc` on Linux), Flutter 3.44+

```bash
# Terminal 1 — backend on :8080
make dev-backend

# Terminal 2 — Flutter UI in Chrome on :3001
make dev-ui
```

Run `make dev` to print these instructions with URLs at any time.

---

## All Make Targets

```bash
make help        # print all targets
```

### Dev (local, no Docker)

| Target | Description |
|--------|-------------|
| `make dev` | Print instructions for running the full local stack |
| `make dev-backend` | Start Go backend on `:8080` (foreground) |
| `make dev-ui` | Start Flutter UI in Chrome on `:3001` (hot reload) |
| `make ui-run` | Same as dev-ui (auto-detects backend URL) |
| `make ui-run-port` | Flutter UI on fixed port 3001 |

### Testing

| Target | Description |
|--------|-------------|
| `make test` | Backend tests + Flutter analysis (CI gate) |
| `make backend-test` | All Go tests with `-race -tags fts5` |
| `make backend-test-v` | Verbose Go test output |
| `make backend-test-cover` | Coverage report → `backend/coverage.html` |
| `make ui-analyze` | Flutter static analysis |
| `make ui-test` | Flutter widget/unit tests |

### Build

| Target | Description |
|--------|-------------|
| `make build` | Backend binary + Flutter web release bundle |
| `make backend-build` | Go binary → `backend/bin/plexreader` |
| `make ui-build` | Flutter web → `ui/build/web/` |

### Docker (production-like)

| Target | Description |
|--------|-------------|
| `make docker-up` | Build images + start stack (foreground) |
| `make docker-up-d` | Build images + start stack (detached) |
| `make docker-down` | Stop containers |
| `make docker-down-v` | Stop containers + wipe database volume |
| `make docker-build` | Build both Docker images |
| `make docker-build-nc` | Build without cache |
| `make docker-logs` | Tail all container logs |
| `make docker-logs-backend` | Tail backend logs only |
| `make docker-logs-ui` | Tail frontend logs only |
| `make docker-ps` | Show running containers |
| `make docker-backend` | Rebuild + restart backend only |
| `make docker-ui` | Rebuild + restart frontend only |
| `make docker-restart` | Restart all without rebuild |
| `make docker-shell-backend` | Shell into running backend container |

### Other

| Target | Description |
|--------|-------------|
| `make tidy` | `go mod tidy` |
| `make proto` | Regenerate Go code from `.proto` (requires `buf`) |
| `make lint-proto` | Lint proto files |
| `make clean` | Remove `bin/`, `build/`, coverage files |
| `make clean-docker` | Remove project Docker images |

---

## Architecture

```
Browser / Flutter Web
        │  Connect-JSON (HTTP POST)
        ▼
nginx (port 3000)
  ├── /plexreader.v1.*  → proxy → backend:8080
  └── /*               → Flutter SPA (index.html)
        │
        ▼
Go Backend (port 8080)
  connect-go: FolderService · FeedService · ArticleService · PreferencesService
        │
        ▼
SQLite WAL (single file)
  + FTS5 virtual table for full-text search
  + Background scheduler (goroutine, semaphore=5)
```

All API contracts are defined in `proto/plexreader/v1/` first. `buf generate` compiles them into Go server stubs, validation code, and an OpenAPI spec. The Flutter frontend speaks Connect-JSON (plain HTTP POST + JSON body) — no generated Dart client required.

See [docs/architecture.md](docs/architecture.md) for the full design including storage schema, auth model, SSRF protection, and request lifecycle.

---

## Documentation

| Document | Description |
|----------|-------------|
| [docs/architecture.md](docs/architecture.md) | System design, storage schema, auth, SSRF, scheduler |
| [docs/installation.md](docs/installation.md) | Docker Compose setup, environment variables, TLS, backups |
| [docs/development.md](docs/development.md) | Repo layout, Make targets, proto workflow, adding endpoints |
| [docs/user-guide.md](docs/user-guide.md) | Using the UI: feeds, keyboard shortcuts, search, preferences |

---

## Configuration

All configuration is via environment variables. See [docs/installation.md](docs/installation.md) for the full reference.

| Variable | Default | Description |
|----------|---------|-------------|
| `PLEXREADER_PORT` | `8080` | Backend listen port |
| `PLEXREADER_DB_PATH` | `./data/plexreader.db` | SQLite database path |
| `PLEXREADER_REFRESH_INTERVAL` | `15m` | Feed refresh interval (`5m`, `1h`, etc.) |
| `PLEXREADER_AUTH_ENABLED` | `false` | Require Bearer token on all requests |
| `PLEXREADER_AUTH_TOKEN` | _(empty)_ | Token value (min 32 chars when auth enabled) |
| `PLEXREADER_ALLOWED_ORIGINS` | `*` | Comma-separated CORS origins |

---

## Contributing

1. Open an issue before starting significant work.
2. Follow proto-first: update `.proto` before Go or Dart code.
3. `make test` must pass before submitting a PR.
4. Write tests for new behaviour; Go backend targets 95%+ coverage.

---

## License

PlexReader is released under the **GNU Lesser General Public License v2.1 or later**. See [LICENSE](LICENSE).

*Not affiliated with Feedly, Inc.*
