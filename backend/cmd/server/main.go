package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"connectrpc.com/connect"
	"connectrpc.com/grpcreflect"
	"github.com/rs/cors"
	"github.com/rs/zerolog"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/plexreader/plexreader/backend/gen/plexreader/v1/plexreaderv1connect"
	"github.com/plexreader/plexreader/backend/internal/feed"
	"github.com/plexreader/plexreader/backend/internal/middleware"
	"github.com/plexreader/plexreader/backend/internal/scheduler"
	"github.com/plexreader/plexreader/backend/internal/service"
	"github.com/plexreader/plexreader/backend/internal/storage"
)

// swaggerJSON is embedded at build time via the swagger.go file in this package.

func main() {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).
		With().Timestamp().Logger()

	cfg := loadConfig()

	// Ensure data directory exists.
	if err := os.MkdirAll(filepath.Dir(cfg.DBPath), 0o755); err != nil {
		logger.Fatal().Err(err).Msg("create data dir")
	}

	// Database.
	db, err := storage.NewDB(cfg.DBPath)
	if err != nil {
		logger.Fatal().Err(err).Msg("open database")
	}
	logger.Info().Str("path", cfg.DBPath).Msg("database ready")

	// Stores.
	folderStore := storage.NewFolderStore(db)
	feedStore := storage.NewFeedStore(db)
	articleStore := storage.NewArticleStore(db)
	prefStore := storage.NewPreferencesStore(db)

	// Feed fetcher.
	fetcher := feed.NewFetcher(30 * time.Second)

	// Scheduler.
	sched := scheduler.New(feedStore, articleStore, prefStore, fetcher, cfg.RefreshInterval, cfg.FetchConcurrency, logger)
	logger.Info().Int("concurrency", sched.Concurrency()).Msg("scheduler configured")
	sched.Start()
	defer sched.Stop()

	// Auth interceptor.
	var authValidator middleware.TokenValidator
	if cfg.AuthEnabled && cfg.AuthToken != "" {
		authValidator = middleware.NewStaticTokenValidator(cfg.AuthToken)
	}
	authInterceptor := middleware.NewAuthInterceptor(cfg.AuthEnabled, authValidator)
	interceptors := connect.WithInterceptors(authInterceptor)

	// Register services.
	mux := http.NewServeMux()

	mux.Handle(plexreaderv1connect.NewFolderServiceHandler(
		service.NewFolderService(folderStore), interceptors,
	))
	mux.Handle(plexreaderv1connect.NewFeedServiceHandler(
		service.NewFeedService(feedStore, folderStore, sched), interceptors,
	))
	mux.Handle(plexreaderv1connect.NewArticleServiceHandler(
		service.NewArticleService(articleStore), interceptors,
	))
	mux.Handle(plexreaderv1connect.NewPreferencesServiceHandler(
		service.NewPreferencesService(prefStore), interceptors,
	))

	// gRPC reflection (for tools like grpcurl).
	reflector := grpcreflect.NewStaticReflector(
		plexreaderv1connect.FolderServiceName,
		plexreaderv1connect.FeedServiceName,
		plexreaderv1connect.ArticleServiceName,
		plexreaderv1connect.PreferencesServiceName,
	)
	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))

	// Limit SQLite to one writer at a time; WAL allows concurrent readers.
	if sqlDB, err := db.DB(); err == nil {
		sqlDB.SetMaxOpenConns(1)
	}

	// Health check — also probes the database.
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if err := db.Raw("SELECT 1").Error; err != nil {
			http.Error(w, "db unavailable", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})

	// Version endpoint.
	mux.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"version":%q}`, Version)
	})

	// Swagger UI — serve the generated spec and a redirect to swagger-ui CDN.
	mux.HandleFunc("/swagger/spec.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(swaggerJSON)
	})
	mux.HandleFunc("/swagger/", func(w http.ResponseWriter, r *http.Request) {
		// Only allow http/https to prevent open-redirect via X-Forwarded-Proto injection.
		scheme := r.Header.Get("X-Forwarded-Proto")
		if scheme != "https" {
			scheme = "http"
		}
		specURL := fmt.Sprintf("%s://%s/swagger/spec.json", scheme, r.Host)
		http.Redirect(w, r,
			"https://petstore.swagger.io/?url="+specURL,
			http.StatusTemporaryRedirect,
		)
	})

	// CORS for Flutter web / Chrome extension.
	// AllowedHeaders must include Connect/gRPC-Web protocol headers.
	// ExposedHeaders must include headers the browser JS needs to read.
	corsHandler := cors.New(cors.Options{
		AllowedOrigins: cfg.AllowedOrigins,
		AllowedMethods: []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{
			"Authorization", "Content-Type",
			"Connect-Protocol-Version", "Connect-Timeout-Ms",
			"Grpc-Timeout", "X-Grpc-Web", "X-User-Agent",
		},
		ExposedHeaders: []string{
			"Grpc-Status", "Grpc-Message", "Grpc-Status-Details-Bin",
			"Connect-Protocol-Version", "Trailer",
		},
		AllowCredentials: false,
	})

	// Limit request bodies to 10 MB to prevent OOM from huge payloads.
	limitedHandler := http.MaxBytesHandler(corsHandler.Handler(mux), 10<<20)

	addr := ":" + cfg.Port
	srv := &http.Server{
		Addr:              addr,
		Handler:           h2c.NewHandler(limitedHandler, &http2.Server{}),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// Graceful shutdown.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info().Str("addr", addr).Msg("server starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("server error")
		}
	}()

	<-shutdown
	logger.Info().Msg("shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error().Err(err).Msg("shutdown error")
	}
	logger.Info().Msg("server stopped")
}
