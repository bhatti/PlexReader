# PlexReader — Architecture

## System Overview

```
┌─────────────────────────────────────────────────────────────────┐
│  Client Layer                                                   │
│                                                                 │
│   ┌──────────────────┐    ┌──────────────────────────────────┐  │
│   │  Chrome Extension│    │  Flutter Web (browser / PWA)     │  │
│   │  (popup.js)      │    │  Riverpod · go_router            │  │
│   └────────┬─────────┘    └─────────────────┬────────────────┘  │
└────────────┼─────────────────────────────────┼──────────────────┘
             │  HTTP/2  (Connect-JSON)          │  HTTP/2  (Connect-JSON)
             ▼                                  ▼
┌─────────────────────────────────────────────────────────────────┐
│  nginx (ui container, port 3000)                                │
│  Serves Flutter web bundle · Proxies /plexreader.v1.* → backend:8080 │
└──────────────────────────────┬──────────────────────────────────┘
                               │  HTTP/2 (Connect / gRPC / gRPC-Web)
                               ▼
┌─────────────────────────────────────────────────────────────────┐
│  Go Backend (port 8080)                                         │
│                                                                 │
│  ┌─────────────┐  ┌─────────────┐  ┌──────────────────────┐    │
│  │FolderService│  │ FeedService │  │   ArticleService     │    │
│  └──────┬──────┘  └──────┬──────┘  └──────────┬───────────┘    │
│         │                │                    │                 │
│  ┌──────┴──────┐  ┌──────┴──────┐  ┌──────────┴───────────┐    │
│  │FolderStore  │  │  FeedStore  │  │    ArticleStore       │    │
│  └──────┬──────┘  └──────┬──────┘  └──────────┬───────────┘    │
│         └────────────────┴──────────────────────┘               │
│                          │  GORM + mattn/go-sqlite3              │
│                          ▼                                      │
│                 ┌──────────────────┐                            │
│                 │  SQLite WAL      │                            │
│                 │  + FTS5 virtual  │                            │
│                 │  table           │                            │
│                 └──────────────────┘                            │
│                                                                 │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  Scheduler (background goroutine)                        │   │
│  │  ticker · semaphore(5) · context cancellation            │   │
│  └──────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

---

## Backend Stack

| Component | Technology | Notes |
|-----------|-----------|-------|
| Language | Go 1.23 | CGO required for SQLite FTS5 |
| RPC framework | connectrpc.com/connect v1 | Single-port gRPC + Connect-JSON + gRPC-Web |
| Database driver | gorm.io/driver/sqlite (mattn/go-sqlite3) | CGO, WAL mode, FTS5 build tag |
| ORM | gorm.io/gorm | Schema migration via `AutoMigrate` |
| Feed parsing | github.com/mmcdole/gofeed | RSS 0.9–2.0, Atom 0.3/1.0, JSON Feed |
| Logging | github.com/rs/zerolog | Structured JSON, console writer in dev |
| ID generation | github.com/oklog/ulid/v2 | Lexicographically sortable, K-sortable |
| CORS | github.com/rs/cors | Required for Flutter web + Chrome extension |
| HTTP/2 | golang.org/x/net/http2 (h2c) | Cleartext HTTP/2, no TLS required |
| Proto validation | buf.build/gen/go/bufbuild/protovalidate | Field-level validation generated from proto annotations |
| gRPC reflection | connectrpc.com/grpcreflect | Enables `grpcurl` introspection |

---

## Frontend Stack

| Component | Technology | Notes |
|-----------|-----------|-------|
| Language | Dart 3 / Flutter 3.44+ | Compiled to JavaScript for web |
| Web bootstrap | `flutter_bootstrap.js` template | Flutter 3.44 pattern; `index.html` loads `flutter_bootstrap.js` with `{{flutter_js}}` / `{{flutter_build_config}}` placeholders |
| State management | flutter_riverpod ^2.5 + riverpod_annotation | `StateNotifier` pattern, code generation |
| Routing | go_router ^14 | Declarative URL-based navigation |
| API protocol | Connect-JSON over `http` package | Protobuf-JSON serialisation, no generated Dart client |
| HTML rendering | flutter_html ^3 (beta) | Article content with safe tag allowlist |
| Images | cached_network_image ^3 | Disk + memory cache for thumbnails |
| File picker | file_picker ^8 | OPML file import from local disk |
| Time formatting | timeago ^3 | Human-readable relative timestamps |
| Theme | Material 3 + Google Fonts | Dark theme by default, `AppColors` constants |

---

## Proto-First Design

All API contracts live in `proto/plexreader/v1/`. No Go or Dart code is written until the proto definition is finalised.

```
proto/plexreader/v1/
├── common.proto        # Shared enums (SortOrder, ViewMode, etc.)
├── folder.proto        # FolderService — CRUD for feed groups
├── feed.proto          # FeedService — subscriptions, OPML, refresh
├── article.proto       # ArticleService — list, mark-read, star, search
└── preferences.proto   # PreferencesService — global reader settings
```

Code generation is driven by `buf.gen.yaml`:

```bash
buf generate   # produces backend/gen/plexreader/v1/ Go stubs + validation
```

Generated output:
- `*.pb.go` — protobuf message types
- `*.pb.validate.go` — protovalidate field validation
- `plexreaderv1connect/*.connect.go` — connect-go service handlers and clients
- `*.swagger.json` — OpenAPI v2 spec (served at `/swagger/spec.json`)

The Flutter frontend manually constructs JSON request bodies matching the proto-JSON wire format and parses JSON responses into hand-written Dart model classes (`ui/lib/models/`). This avoids a Dart protobuf code-generation dependency while keeping the wire format consistent.

---

## Storage Layer

### Schema

All tables are created via GORM `AutoMigrate` at startup. The primary key for every record is a ULID string — lexicographically sortable, making cursor-based pagination trivial.

```
folders
  id TEXT PRIMARY KEY        -- ULID
  name TEXT NOT NULL
  parent_id TEXT
  position INTEGER

feeds
  id TEXT PRIMARY KEY        -- ULID
  title TEXT
  xml_url TEXT UNIQUE        -- deduplicated on subscribe
  html_url TEXT
  folder_id TEXT (index)
  refresh_interval_seconds INTEGER
  last_fetched_at DATETIME (index)
  last_error TEXT
  error_count INTEGER

articles
  id TEXT PRIMARY KEY        -- ULID
  feed_id TEXT (index)
  title TEXT
  link TEXT
  content TEXT               -- full HTML body
  summary TEXT
  author TEXT
  published_at DATETIME (index)
  guid TEXT
  guid_feed_id TEXT          -- composite unique: (guid, feed_id)
  thumbnail_url TEXT
  is_read BOOLEAN (index)
  is_starred BOOLEAN (index)
  is_saved_for_later BOOLEAN (index)
  read_at DATETIME
  created_at DATETIME

user_preferences             -- single row, id=1
  start_page TEXT
  default_view TEXT
  default_sort TEXT
  hide_read_articles BOOLEAN
  global_refresh_interval_seconds INTEGER
  retention_days INTEGER
  theme TEXT
```

### FTS5 Full-Text Search

A virtual FTS5 table mirrors the `articles` title and content columns:

```sql
CREATE VIRTUAL TABLE articles_fts USING fts5(
  title, content,
  content='articles', content_rowid='rowid',
  tokenize='porter ascii'
);
```

Insert/update/delete triggers keep the FTS index in sync automatically. Search queries use the `MATCH` operator with relevance ranking via `bm25()`:

```sql
SELECT a.* FROM articles a
JOIN articles_fts f ON a.rowid = f.rowid
WHERE articles_fts MATCH ?
ORDER BY bm25(articles_fts) ASC
LIMIT ? OFFSET ?;
```

### Cursor-Based Pagination

Article list endpoints use opaque cursor tokens (base64-encoded `published_at` + `id` pairs) rather than `OFFSET` pagination. This keeps page-flip latency constant as the article table grows and avoids duplicates when new articles are inserted between pages.

```go
// pagination.go — encode/decode cursor
type Cursor struct {
    PublishedAt time.Time
    ID          string
}
```

### WAL Mode

The database is opened with `PRAGMA journal_mode=WAL` and `PRAGMA synchronous=NORMAL`. WAL mode allows concurrent reads during the background scheduler's writes without blocking the API handlers.

---

## Auth

Auth is opt-in and disabled by default (`PLEXREADER_AUTH_ENABLED=false`). When enabled, every Connect/gRPC request must carry a `Bearer` token in the `Authorization` header.

```
Client → Authorization: Bearer <token> → AuthInterceptor → service handler
```

The interceptor is a `connect.Interceptor` wrapping both unary and streaming handlers:

```go
type TokenValidator interface {
    Validate(token string) error
}
```

The only built-in implementation is `StaticTokenValidator`, which performs a constant-time comparison (`crypto/subtle`) against the token configured via `PLEXREADER_AUTH_TOKEN`. The interface is intentionally small so future implementors can drop in JWT verification, OIDC introspection, or API-key lookup against the database without changing any service code.

---

## Background Scheduler

The scheduler runs a single goroutine that wakes on a `time.Ticker` (default 15 minutes). On each cycle it:

1. Queries `FeedStore.ListDueForRefresh` — feeds whose `last_fetched_at` is older than their `refresh_interval_seconds`.
2. Dispatches up to **5 concurrent** fetch goroutines, controlled by a buffered-channel semaphore.
3. For each feed, calls `Fetcher.Fetch` which uses HTTP conditional requests (`If-None-Match` / `If-Modified-Since`) to avoid re-downloading unchanged content.
4. Bulk-inserts new articles with `ArticleStore.BulkCreate`. The `(guid, feed_id)` unique constraint silently deduplicates items already present.
5. After all feeds complete, runs the retention sweep: deletes articles older than `UserPreferences.RetentionDays`.

Shutdown is coordinated via `context.WithCancel` — `sched.Stop()` cancels the context, in-flight fetches respect the context deadline, and the caller blocks on `sync.WaitGroup` until the goroutine exits cleanly.

---

## SSRF Protection

Two complementary layers in `internal/feed/fetcher.go` prevent server-side request forgery.

### Pre-flight URL validation (`validateFeedURL`)

Called before any HTTP request is made:

1. Rejects non-`http`/`https` schemes (blocks `file://`, `ftp://`, etc.).
2. Rejects `localhost` and `*.localhost` hostnames.
3. Resolves the hostname via DNS and rejects IPs in:
   - IPv4 loopback (`127.0.0.0/8`)
   - IPv4 link-local (`169.254.0.0/16`)
   - IPv4 private ranges (RFC 1918: `10/8`, `172.16/12`, `192.168/16`)
   - IPv6 loopback and link-local
   - IPv6 unique-local (`fc00::/7`)
4. **DNS failures are now errors (fail-closed).** A failed DNS lookup returns an error rather than allowing the request through.

### DNS-rebinding protection (`safeDialContext`)

A pre-flight DNS check alone is not sufficient. A malicious DNS server can return a public IP for the validation lookup and a private IP for the actual TCP connection (DNS rebinding). `safeDialContext` closes this window by re-resolving and re-validating the IP **at TCP-connect time**:

```go
// Installed as http.Transport.DialContext — runs on every TCP connection
func safeDialContext(ctx context.Context, network, addr string) (net.Conn, error) {
    host, port, _ := net.SplitHostPort(addr)
    addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
    // reject private IPs, cloud metadata endpoints (169.254.169.254, 100.100.100.200)
    // normalises IPv4-mapped IPv6 via ip.To4() before checking
    dialer := &net.Dialer{}
    return dialer.DialContext(ctx, network, net.JoinHostPort(addrs[0].IP.String(), port))
}
```

Cloud metadata endpoints (`169.254.169.254/32` for AWS/GCP/Azure, `100.100.100.200/32` for Alibaba Cloud) are blocked by an explicit allowlist checked at both layers.

### HTML Sanitisation

All article HTML content is passed through [bluemonday](https://github.com/microcosm-cc/bluemonday) `UGCPolicy` before being stored in SQLite. This strips XSS payloads (event handlers, `javascript:` URLs, dangerous tags) from feed content while preserving safe formatting.

---

## Package Dependency Diagram

```
cmd/server
    │
    ├── internal/middleware     (connect interceptor — auth)
    ├── internal/scheduler      (background refresh loop)
    ├── internal/service        (connect service implementations)
    │       └── internal/storage    (GORM stores + models)
    ├── internal/feed           (HTTP fetcher, OPML parser, feed parser)
    └── gen/plexreader/v1/      (generated proto types + connect stubs)
```

Services depend on storage interfaces (not concrete types), enabling unit tests to substitute in-memory fakes without a real SQLite database.

---

## Request Lifecycle (Example: List Articles)

```
Flutter UI
  │  POST /plexreader.v1.ArticleService/ListArticles
  │  Content-Type: application/connect+json
  ▼
h2c.Handler (HTTP/2 cleartext)
  ▼
cors.Handler (allow Flutter web origin)
  ▼
http.MaxBytesHandler (10 MB body limit)
  ▼
connect-go router → ArticleServiceHandler
  ▼
AuthInterceptor.WrapUnary (verify Bearer token if auth enabled)
  ▼
ArticleService.ListArticles
  ▼
ArticleStore.List (GORM query, cursor pagination)
  ▼
SQLite WAL database
  ▼
protobuf-JSON response → Flutter UI
```

---

## Deployment Topology

For production use, the recommended topology is:

```
Internet → reverse proxy (nginx / Caddy with TLS) → docker-compose stack
                                                         ├── frontend:3000
                                                         └── backend:8080
```

The backend container does not terminate TLS — the reverse proxy handles certificates. The frontend nginx container proxies all `/plexreader.v1.*` calls to the backend; the browser only needs to reach one origin (port 443 on your domain).
