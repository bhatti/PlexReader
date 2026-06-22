package main

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type config struct {
	Port             string
	DBPath           string
	RefreshInterval  time.Duration
	FetchConcurrency int
	AuthEnabled      bool
	AuthToken        string
	AllowedOrigins   []string
}

func loadConfig() config {
	cfg := config{
		Port:        envOr("PLEXREADER_PORT", "8080"),
		DBPath:      envOr("PLEXREADER_DB_PATH", "./data/plexreader.db"),
		AuthEnabled: os.Getenv("PLEXREADER_AUTH_ENABLED") == "true",
		AuthToken:   os.Getenv("PLEXREADER_AUTH_TOKEN"),
	}

	if v := os.Getenv("PLEXREADER_FETCH_CONCURRENCY"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.FetchConcurrency = n
		}
	}

	// Fail fast: if auth is enabled, require a token of at least 32 chars.
	if cfg.AuthEnabled && len(cfg.AuthToken) < 32 {
		log.Fatal("PLEXREADER_AUTH_TOKEN must be at least 32 characters when PLEXREADER_AUTH_ENABLED=true")
	}

	interval, err := time.ParseDuration(envOr("PLEXREADER_REFRESH_INTERVAL", "15m"))
	if err != nil {
		interval = 15 * time.Minute
	}
	cfg.RefreshInterval = interval

	rawOrigins := os.Getenv("PLEXREADER_ALLOWED_ORIGINS")
	if rawOrigins != "" {
		cfg.AllowedOrigins = strings.Split(rawOrigins, ",")
	} else {
		cfg.AllowedOrigins = []string{"*"}
	}

	return cfg
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
