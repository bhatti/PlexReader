# PlexReader — Developer Guide

## Table of Contents

- [Repository Structure](#repository-structure)
- [Make Targets Reference](#make-targets-reference)
- [Proto Workflow](#proto-workflow)
- [Running Tests](#running-tests)
- [Adding a New API Endpoint](#adding-a-new-api-endpoint)
- [Code Conventions](#code-conventions)
- [Docker Build](#docker-build)
- [Debugging Tips](#debugging-tips)

---

## Repository Structure

```
plexreader/
├── proto/                        # Source of truth for all API contracts
│   └── plexreader/v1/
│       ├── common.proto          # Shared enums: SortOrder, ViewMode, StartPage
│       ├── folder.proto          # FolderService — CRUD for feed groups
│       ├── feed.proto            # FeedService — subscriptions, OPML, refresh
│       ├── article.proto         # ArticleService — list, search, mark-read, star
│       └── preferences.proto     # PreferencesService — global reader settings
│
├── backend/                      # Go backend server
│   ├── cmd/
│   │   └── server/
│   │       ├── main.go           # Entry point: server wiring, graceful shutdown
│   │       ├── config.go         # Environment-variable configuration loading
│   │       └── swagger.go        # Embeds generated swagger.json at build time
│   ├── gen/                      # Generated code — DO NOT EDIT BY HAND
│   │   └── plexreader/v1/
│   │       ├── *.pb.go           # Protobuf message types
│   │       ├── *.pb.validate.go  # protovalidate field validation
│   │       └── plexreaderv1connect/
│   │           └── *.connect.go  # connect-go service stubs and client interfaces
│   ├── internal/
│   │   ├── feed/
│   │   │   ├── fetcher.go        # HTTP fetcher with SSRF protection + conditional requests
│   │   │   ├── parser.go         # gofeed wrapper, extracts thumbnails from media:* tags
│   │   │   ├── opml.go           # OPML 1.0 import and export
│   │   │   ├── parser_test.go
│   │   │   └── opml_test.go
│   │   ├── middleware/
│   │   │   ├── auth.go           # connect-go interceptor: TokenValidator interface
│   │   │   └── auth_test.go
│   │   ├── scheduler/
│   │   │   └── scheduler.go      # Background refresh loop + retention cleanup
│   │   ├── service/
│   │   │   ├── feed_service.go   # FeedService implementation
│   │   │   ├── folder_service.go # FolderService implementation
│   │   │   ├── article_service.go# ArticleService implementation
│   │   │   ├── preferences_service.go
│   │   │   ├── convert.go        # Proto ↔ storage model conversion helpers
│   │   │   ├── errors.go         # Connect error code mapping
│   │   │   └── service_test.go
│   │   └── storage/
│   │       ├── models.go         # GORM model structs (Folder, Feed, Article, UserPreferences)
│   │       ├── db.go             # SQLite open + AutoMigrate + PRAGMA setup
│   │       ├── feed_store.go     # FeedStore interface + GORM implementation
│   │       ├── folder_store.go   # FolderStore interface + GORM implementation
│   │       ├── article_store.go  # ArticleStore: list, bulk create, FTS5 search
│   │       ├── preferences_store.go
│   │       ├── pagination.go     # Cursor encode/decode helpers
│   │       └── storage_test.go
│   ├── go.mod
│   ├── go.sum
│   └── Dockerfile
│
├── ui/                           # Flutter frontend
│   ├── lib/
│   │   ├── main.dart             # App entry point, ProviderScope
│   │   ├── app.dart              # GoRouter configuration, top-level routes
│   │   ├── models/               # Dart model classes (hand-written, match proto-JSON)
│   │   │   ├── article.dart
│   │   │   ├── feed.dart
│   │   │   ├── folder.dart
│   │   │   └── preferences.dart
│   │   ├── providers/            # Riverpod StateNotifier providers
│   │   │   ├── article_provider.dart
│   │   │   ├── feed_provider.dart
│   │   │   ├── folder_provider.dart
│   │   │   ├── preferences_provider.dart
│   │   │   ├── navigation_provider.dart
│   │   │   └── api_client_provider.dart
│   │   ├── services/             # HTTP calls to the Connect-JSON API
│   │   │   ├── api_client.dart   # Base client: base URL, auth header injection
│   │   │   ├── article_service.dart
│   │   │   ├── feed_service.dart
│   │   │   ├── folder_service.dart
│   │   │   └── preferences_service.dart
│   │   ├── screens/              # Full-page widgets (one per route)
│   │   │   ├── app_shell.dart    # Persistent layout: sidebar + content area
│   │   │   ├── today_screen.dart
│   │   │   ├── feed_screen.dart
│   │   │   ├── folder_screen.dart
│   │   │   ├── starred_screen.dart
│   │   │   ├── saved_screen.dart
│   │   │   ├── search_screen.dart
│   │   │   ├── recently_read_screen.dart
│   │   │   ├── all_articles_screen.dart
│   │   │   └── preferences_screen.dart
│   │   ├── widgets/              # Reusable UI components
│   │   │   ├── article_card_magazine.dart
│   │   │   ├── article_card_title.dart
│   │   │   ├── article_card_grid.dart
│   │   │   ├── article_list.dart
│   │   │   ├── article_detail.dart
│   │   │   ├── sidebar.dart
│   │   │   ├── keyboard_shortcuts.dart
│   │   │   ├── add_feed_dialog.dart
│   │   │   ├── import_opml_dialog.dart
│   │   │   ├── create_folder_dialog.dart
│   │   │   ├── unread_badge.dart
│   │   │   ├── empty_state.dart
│   │   │   └── error_state.dart
│   │   └── theme/
│   │       └── app_theme.dart    # AppColors constants, ThemeData
│   ├── chrome_extension/         # Browser extension (Manifest V3)
│   │   ├── manifest.json
│   │   ├── popup.html
│   │   ├── popup.js
│   │   ├── background.js
│   │   └── content_script.js
│   ├── pubspec.yaml
│   ├── nginx.conf                # nginx config for the frontend Docker container
│   └── Dockerfile
│
├── buf.yaml                      # buf module config
├── buf.gen.yaml                  # buf code generation config
├── buf.lock                      # buf dependency lockfile
├── docker-compose.yaml
├── Makefile
└── .env.example
```

---

## Make Targets Reference

### Dev (local, no Docker)

| Target | Description |
|--------|-------------|
| `make help` | Print all targets with descriptions |
| `make dev` | Print instructions for running the full local stack |
| `make dev-backend` | Start Go backend on `:8080` (foreground, hot-reloadable via `go run`) |
| `make dev-ui` | Start Flutter UI in Chrome on `:3001` (hot reload) |
| `make backend-run` | Alias for `dev-backend` |
| `make ui-run` | Flutter UI with auto-detected backend URL |
| `make ui-run-port` | Flutter UI on fixed port 3001 |

### Testing

| Target | Description |
|--------|-------------|
| `make test` | Backend tests + Flutter analysis — the CI gate |
| `make backend-test` | All Go tests with `-race -tags fts5` |
| `make backend-test-v` | Verbose Go test output |
| `make backend-test-cover` | Coverage report → `backend/coverage.html` |
| `make ui-analyze` | Flutter static analysis (used in CI) |
| `make ui-test` | Flutter widget/unit tests |

### Build

| Target | Description |
|--------|-------------|
| `make build` | Backend binary + Flutter web release bundle |
| `make backend-build` | Go binary → `backend/bin/plexreader` |
| `make ui-build` | Flutter web → `ui/build/web/` |
| `make ui-deps` | Run `flutter pub get` |
| `make tidy` | Run `go mod tidy` |

### Docker (production-like)

| Target | Description |
|--------|-------------|
| `make docker-up` | Build images + start stack (foreground) |
| `make docker-up-d` | Build images + start stack (detached) |
| `make docker-down` | Stop and remove containers |
| `make docker-down-v` | Stop containers + wipe named volumes (resets database) |
| `make docker-build` | Build both Docker images |
| `make docker-build-nc` | Build without layer cache |
| `make docker-logs` | Tail logs from all services |
| `make docker-logs-backend` | Tail backend logs only |
| `make docker-logs-ui` | Tail frontend logs only |
| `make docker-ps` | Show running containers |
| `make docker-restart` | Restart all services without rebuilding |
| `make docker-backend` | Rebuild + restart backend only |
| `make docker-ui` | Rebuild + restart frontend only |
| `make docker-shell-backend` | Open a shell in the running backend container |

### Proto & misc

| Target | Description |
|--------|-------------|
| `make proto` | Run `buf generate` to regenerate Go code from proto files |
| `make lint-proto` | Lint proto files with `buf lint` |
| `make clean` | Remove `backend/bin/`, `ui/build/`, coverage files |
| `make clean-docker` | Remove project Docker images |

---

## Proto Workflow

PlexReader follows a **proto-first** discipline. All data shapes and service contracts are defined in `.proto` files before any Go or Dart code is written.

### Edit a proto file

```bash
vim proto/plexreader/v1/article.proto
```

### Validate and regenerate

```bash
# Lint for style violations
make lint-proto

# Check that you haven't broken any existing clients
make check-proto-breaking

# Regenerate Go code
make proto
```

The generated files land in `backend/gen/plexreader/v1/`. Never edit files in `gen/` — they are overwritten on every `buf generate` run.

### buf.gen.yaml overview

```yaml
# buf.gen.yaml
version: v2
plugins:
  - remote: buf.build/protocolbuffers/go      # *.pb.go message types
  - remote: buf.build/connectrpc/go           # plexreaderv1connect/*.connect.go
  - remote: buf.build/grpc-ecosystem/gateway  # gRPC-gateway HTTP bindings
  - remote: buf.build/grpc-ecosystem/openapiv2 # swagger.json spec
  - remote: buf.build/bufbuild/validate-go    # *.pb.validate.go field validation
```

### Adding protovalidate constraints

Use the `buf.validate` option annotations in the `.proto` file:

```protobuf
import "buf/validate/validate.proto";

message CreateFeedRequest {
  string xml_url = 1 [
    (buf.validate.field).string = {min_len: 1, uri: true}
  ];
}
```

After `buf generate`, the generated `*.pb.validate.go` file contains `Validate()` methods. The connect-go interceptors call these automatically for every request.

---

## Running Tests

### Backend

The backend tests require CGO and the `fts5` build tag (for SQLite full-text search):

```bash
make backend-test
# expands to:
# cd backend && CGO_ENABLED=1 go test -tags fts5 -race -timeout 60s ./...
```

For a faster feedback loop during development, run a single package:

```bash
cd backend && CGO_ENABLED=1 go test -tags fts5 -race ./internal/storage/...
cd backend && CGO_ENABLED=1 go test -tags fts5 -race ./internal/service/...
cd backend && CGO_ENABLED=1 go test -tags fts5 -race ./internal/feed/...
```

Storage tests use an in-memory SQLite database (`:memory:`). Service tests use the real storage layer with an in-memory database. There are no mocks — tests exercise the full stack down to SQLite.

### Frontend

```bash
make frontend-test
# expands to: cd ui && flutter test
```

Flutter widget tests use `flutter_test` and `flutter_riverpod`'s `ProviderContainer` for dependency injection in tests. HTTP calls are stubbed using the `http` package's mock client.

### CI pipeline

```bash
make test   # backend-test + ui-analyze
```

`ui-analyze` is used in CI because widget tests require a display. Static analysis catches type errors, deprecated API usage, and lint violations.

---

## Adding a New API Endpoint

Follow these steps to add a new endpoint end-to-end. The example adds a `GetArticleStats` RPC that returns per-feed article counts.

### 1. Define the proto

Edit `proto/plexreader/v1/article.proto`:

```protobuf
message ArticleStats {
  string feed_id = 1;
  int32 total_count = 2;
  int32 unread_count = 3;
}

message GetArticleStatsRequest {
  string feed_id = 1 [
    (google.api.field_behavior) = REQUIRED,
    (buf.validate.field).string.min_len = 1
  ];
}

service ArticleService {
  // ... existing RPCs ...

  rpc GetArticleStats(GetArticleStatsRequest) returns (ArticleStats) {
    option (google.api.http) = {
      get: "/v1/articles/stats"
    };
  }
}
```

### 2. Regenerate code

```bash
make proto
```

This adds `GetArticleStats` to `ArticleServiceHandler` interface in `backend/gen/`.

### 3. Implement the storage layer

Add a method to `ArticleStore` interface in `backend/internal/storage/article_store.go`:

```go
type ArticleStore interface {
    // ... existing methods ...
    GetStats(ctx context.Context, feedID string) (total, unread int64, err error)
}
```

Implement it in the same file:

```go
func (s *articleStore) GetStats(ctx context.Context, feedID string) (int64, int64, error) {
    var total, unread int64
    if err := s.db.WithContext(ctx).Model(&Article{}).
        Where("feed_id = ?", feedID).Count(&total).Error; err != nil {
        return 0, 0, err
    }
    if err := s.db.WithContext(ctx).Model(&Article{}).
        Where("feed_id = ? AND is_read = false", feedID).Count(&unread).Error; err != nil {
        return 0, 0, err
    }
    return total, unread, nil
}
```

### 4. Implement the service

Add to `ArticleService` in `backend/internal/service/article_service.go`:

```go
func (s *ArticleService) GetArticleStats(
    ctx context.Context,
    req *connect.Request[pb.GetArticleStatsRequest],
) (*connect.Response[pb.ArticleStats], error) {
    if err := req.Msg.Validate(); err != nil {
        return nil, connect.NewError(connect.CodeInvalidArgument, err)
    }
    total, unread, err := s.articleStore.GetStats(ctx, req.Msg.FeedId)
    if err != nil {
        return nil, mapError(err)
    }
    return connect.NewResponse(&pb.ArticleStats{
        FeedId:      req.Msg.FeedId,
        TotalCount:  int32(total),
        UnreadCount: int32(unread),
    }), nil
}
```

### 5. Add a Flutter service method

Add to `ui/lib/services/article_service.dart`:

```dart
Future<ArticleStats> getArticleStats(String feedId) async {
  final response = await _client.post(
    '/plexreader.v1.ArticleService/GetArticleStats',
    body: jsonEncode({'feed_id': feedId}),
  );
  _client.checkStatus(response);
  return ArticleStats.fromJson(jsonDecode(response.body) as Map<String, dynamic>);
}
```

Add the `ArticleStats` model to `ui/lib/models/article.dart`:

```dart
class ArticleStats {
  final String feedId;
  final int totalCount;
  final int unreadCount;

  const ArticleStats({
    required this.feedId,
    required this.totalCount,
    required this.unreadCount,
  });

  factory ArticleStats.fromJson(Map<String, dynamic> json) => ArticleStats(
    feedId: json['feed_id'] as String? ?? '',
    totalCount: json['total_count'] as int? ?? 0,
    unreadCount: json['unread_count'] as int? ?? 0,
  );
}
```

### 6. Add a Riverpod provider

In `ui/lib/providers/article_provider.dart`:

```dart
final articleStatsProvider = FutureProvider.family<ArticleStats, String>(
  (ref, feedId) => ref.watch(articleServiceProvider).getArticleStats(feedId),
);
```

### 7. Write tests

```bash
# Backend — add to backend/internal/service/service_test.go
# Frontend — add to ui/test/

make backend-test
make ui-test
```

---

## Code Conventions

### Go Backend

**IDs**: All record IDs are ULIDs (`github.com/oklog/ulid/v2`). ULIDs are lexicographically sortable, which enables efficient cursor-based pagination on string primary keys.

```go
import "github.com/oklog/ulid/v2"

id := ulid.Make().String()   // correct
// uuid.New().String()       // never — use ULID
```

**Logging**: Use `zerolog`. Always include relevant context fields; never use `fmt.Printf` or `log.Println`.

```go
logger.Info().
    Str("feed_id", f.ID).
    Int("articles", count).
    Msg("feed refreshed")

logger.Error().Err(err).Str("feed_id", f.ID).Msg("refresh failed")
```

**Error handling**: Map storage errors to connect error codes in `service/errors.go`. Return `connect.NewError(connect.CodeNotFound, err)` for missing records, `CodeInvalidArgument` for validation failures, `CodeInternal` for unexpected errors.

**GORM**: Use `WithContext(ctx)` on every query so the scheduler's context cancellation propagates to in-flight database queries:

```go
var feed Feed
err := s.db.WithContext(ctx).First(&feed, "id = ?", id).Error
```

**Interfaces**: Each store is defined as an interface in its own file. Implementations are unexported structs. This keeps service code testable without a real database.

### Flutter Frontend

**State management**: All mutable state lives in Riverpod `StateNotifier` classes. Widgets are `ConsumerWidget` or `ConsumerStatefulWidget`; they call `ref.watch` to subscribe and `ref.read` to invoke mutations.

```dart
// Provider definition
final feedsProvider = StateNotifierProvider<FeedsNotifier, AsyncValue<List<Feed>>>(
  (ref) => FeedsNotifier(ref.watch(feedServiceProvider)),
);

// In a widget
final feeds = ref.watch(feedsProvider);
feeds.when(
  data: (list) => FeedList(feeds: list),
  loading: () => const CircularProgressIndicator(),
  error: (e, st) => ErrorState(message: e.toString()),
);
```

**API calls**: All HTTP calls go through `ApiClient` (`ui/lib/services/api_client.dart`), which injects the base URL (from `--dart-define=API_BASE_URL`) and the auth token from shared preferences. Service classes (`FeedService`, `ArticleService`, etc.) use `ApiClient` and return typed model objects.

**Models**: Dart models are hand-written. They implement `fromJson` factory constructors and `toJson` methods that match the proto-JSON field names (snake_case).

**Error display**: Use `ErrorState` widget for full-page errors and `SnackBar` for transient mutations. Never swallow exceptions silently.

**Null safety**: All models use non-nullable fields where possible. Use `??` default values in `fromJson` to handle missing fields gracefully when the API adds new optional fields.

---

## Docker Build

### Build images locally

```bash
make docker-build
# Produces:
#   plexreader-backend:latest
#   plexreader-ui:latest
```

Both images are multi-stage builds. Both Dockerfiles use the **repo root** as the build context (`context: .` in docker-compose.yaml), so all `COPY` paths are prefixed with `backend/` or `ui/`.

**backend/Dockerfile**:
1. `golang:1.23-alpine` with `gcc musl-dev sqlite-dev` — compiles the binary with `CGO_ENABLED=1 -tags fts5`.
2. `alpine:3.20` with `ca-certificates sqlite-libs tzdata wget` — copies just the binary; no Go toolchain in the final image. `wget` is used for the Docker `HEALTHCHECK`.

**ui/Dockerfile**:
1. `ghcr.io/cirruslabs/flutter:stable` — runs `flutter build web --release --dart-define=API_BASE_URL=`. The empty `API_BASE_URL` means all Connect-JSON requests go to the same origin; nginx handles routing.
2. `nginx:1.27-alpine` — serves the built web bundle with the project's `nginx.conf`.

### Run the stack

```bash
make docker-up-d   # build + start detached
make docker-down   # stop
make docker-down-v # stop + wipe database volume (full reset)
```

Or run just the backend image directly:

```bash
docker run -p 8080:8080 \
  -e PLEXREADER_DB_PATH=/data/plexreader.db \
  -v plexreader-data:/data \
  plexreader-backend:latest
```

---

## Debugging Tips

### Inspect the API with grpcurl

The backend exposes gRPC reflection, so you can explore and call any RPC without a client:

```bash
# List all services
grpcurl -plaintext localhost:8080 list

# List methods on FeedService
grpcurl -plaintext localhost:8080 list plexreader.v1.FeedService

# Call ListFeeds
grpcurl -plaintext -d '{}' localhost:8080 plexreader.v1.FeedService/ListFeeds
```

### Browse the REST API via Swagger UI

Navigate to **http://localhost:8080/swagger/** in your browser. You are redirected to the Swagger UI CDN with the spec URL pre-loaded. All endpoints can be tried interactively.

### Check the raw SQLite database

```bash
sqlite3 ./data/plexreader.db

# Useful queries:
.tables
SELECT COUNT(*) FROM articles;
SELECT COUNT(*) FROM articles WHERE is_read = 0;
SELECT title, last_fetched_at, last_error FROM feeds ORDER BY last_fetched_at DESC;
SELECT * FROM user_preferences;
```

### Enable verbose Flutter logging

Add `--verbose` to any `flutter` command:

```bash
cd ui && flutter run -d chrome --verbose --dart-define=API_BASE_URL=http://localhost:8080
```

Network requests are logged to the console. Check the browser DevTools Network tab to inspect Connect-JSON request/response payloads.

### Tail backend logs

```bash
# Docker Compose
make docker-logs            # all services
make docker-logs-backend    # backend only

# Local run — logs go to stderr with zerolog console formatting
make dev-backend 2>&1 | tee /tmp/plexreader.log
```
