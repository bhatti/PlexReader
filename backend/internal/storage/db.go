package storage

import (
	"fmt"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewDB opens (or creates) the SQLite database, runs auto-migration, sets up
// FTS5, and seeds default preferences.
func NewDB(dbPath string) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=on", dbPath)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := db.AutoMigrate(
		&Folder{},
		&Feed{},
		&Article{},
		&UserPreferences{},
	); err != nil {
		return nil, fmt.Errorf("auto migrate: %w", err)
	}

	// FTS5 is optional — SQLite builds without it are supported; full-text
	// search will degrade gracefully (queries return no results).
	if err := setupFTS5(db); err != nil {
		log.Printf("WARNING: FTS5 not available, full-text search disabled: %v", err)
	}

	if err := setupIndexes(db); err != nil {
		return nil, fmt.Errorf("setup indexes: %w", err)
	}

	if err := seedDefaults(db); err != nil {
		return nil, fmt.Errorf("seed defaults: %w", err)
	}

	return db, nil
}

// setupFTS5 creates the full-text search virtual table and sync triggers.
// Safe to call on an existing database — uses IF NOT EXISTS.
func setupFTS5(db *gorm.DB) error {
	stmts := []string{
		`CREATE VIRTUAL TABLE IF NOT EXISTS articles_fts
		 USING fts5(title, content, summary, content=articles, content_rowid=rowid)`,

		`CREATE TRIGGER IF NOT EXISTS articles_ai AFTER INSERT ON articles BEGIN
		   INSERT INTO articles_fts(rowid, title, content, summary)
		   VALUES (new.rowid, new.title, new.content, new.summary);
		 END`,

		`CREATE TRIGGER IF NOT EXISTS articles_ad AFTER DELETE ON articles BEGIN
		   INSERT INTO articles_fts(articles_fts, rowid, title, content, summary)
		   VALUES ('delete', old.rowid, old.title, old.content, old.summary);
		 END`,

		`CREATE TRIGGER IF NOT EXISTS articles_au AFTER UPDATE ON articles BEGIN
		   INSERT INTO articles_fts(articles_fts, rowid, title, content, summary)
		   VALUES ('delete', old.rowid, old.title, old.content, old.summary);
		   INSERT INTO articles_fts(rowid, title, content, summary)
		   VALUES (new.rowid, new.title, new.content, new.summary);
		 END`,
	}
	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			return fmt.Errorf("exec FTS5 statement: %w", err)
		}
	}
	return nil
}

// setupIndexes creates composite indexes for the most common article queries.
// Uses IF NOT EXISTS so it is safe to re-run on every startup.
func setupIndexes(db *gorm.DB) error {
	indexes := []string{
		// Unread listing per feed — the most frequent query pattern.
		`CREATE INDEX IF NOT EXISTS idx_articles_feed_unread
		 ON articles(feed_id, is_read, published_at DESC)`,

		// Mark-all-read by feed (WHERE feed_id=? AND is_read=false).
		`CREATE INDEX IF NOT EXISTS idx_articles_feed_isread
		 ON articles(feed_id, is_read)`,

		// Mark-all-read by folder (subquery: feed_id IN (SELECT id FROM feeds WHERE folder_id=?)).
		`CREATE INDEX IF NOT EXISTS idx_feeds_folder
		 ON feeds(folder_id)`,

		// Global unread listing / sort by published_at.
		`CREATE INDEX IF NOT EXISTS idx_articles_isread_published
		 ON articles(is_read, published_at DESC)`,

		// Recently Read view — read_at not null, sorted newest first.
		`CREATE INDEX IF NOT EXISTS idx_articles_readat
		 ON articles(read_at DESC) WHERE read_at IS NOT NULL`,

		// Starred / saved-for-later fast lookup.
		`CREATE INDEX IF NOT EXISTS idx_articles_starred_published
		 ON articles(is_starred, published_at DESC)`,

		`CREATE INDEX IF NOT EXISTS idx_articles_saved_published
		 ON articles(is_saved_for_later, published_at DESC)`,

		// Today screen — recent articles across all feeds.
		`CREATE INDEX IF NOT EXISTS idx_articles_published
		 ON articles(published_at DESC)`,
	}
	for _, stmt := range indexes {
		if err := db.Exec(stmt).Error; err != nil {
			return fmt.Errorf("create index: %w", err)
		}
	}
	return nil
}

// seedDefaults ensures the single UserPreferences row exists.
func seedDefaults(db *gorm.DB) error {
	var count int64
	if err := db.Model(&UserPreferences{}).Count(&count).Error; err != nil {
		return fmt.Errorf("count preferences: %w", err)
	}
	if count == 0 {
		return db.Create(&UserPreferences{ID: 1}).Error
	}
	return nil
}
