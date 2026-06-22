// SPDX-License-Identifier: LGPL-2.1-or-later
package scheduler_test

import (
	"context"
	"testing"
	"time"

	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"

	"github.com/plexreader/plexreader/backend/internal/scheduler"
	"github.com/plexreader/plexreader/backend/internal/storage"
)

func testDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := storage.NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB: %v", err)
	}
	return db
}

// insertOldReadArticle bypasses the store API to set created_at in the past.
func insertOldReadArticle(t *testing.T, db *gorm.DB, feedID string, daysAgo int) string {
	t.Helper()
	id := ulid.Make().String()
	createdAt := time.Now().AddDate(0, 0, -daysAgo)
	err := db.Exec(
		`INSERT INTO articles (id, feed_id, guid, guid_feed_id, title, is_read, created_at) VALUES (?, ?, ?, ?, ?, true, ?)`,
		id, feedID, id, feedID, "old article", createdAt,
	).Error
	if err != nil {
		t.Fatalf("insert old article: %v", err)
	}
	return id
}

func insertFeed(t *testing.T, db *gorm.DB) string {
	t.Helper()
	id := ulid.Make().String()
	err := db.Exec(
		`INSERT INTO feeds (id, title, xml_url, refresh_interval_seconds) VALUES (?, ?, ?, ?)`,
		id, "Test Feed", "http://test.example/rss", 900,
	).Error
	if err != nil {
		t.Fatalf("insert feed: %v", err)
	}
	return id
}

func articleExists(t *testing.T, db *gorm.DB, id string) bool {
	t.Helper()
	var count int64
	db.Raw("SELECT COUNT(*) FROM articles WHERE id = ?", id).Scan(&count)
	return count > 0
}

// TestRetention_DeletesArticlesOlderThan90Days verifies that read articles
// older than 90 days are removed and newer ones are kept.
func TestRetention_DeletesArticlesOlderThan90Days(t *testing.T) {
	db := testDB(t)
	feedStore := storage.NewFeedStore(db)
	articleStore := storage.NewArticleStore(db)
	prefStore := storage.NewPreferencesStore(db)
	ctx := context.Background()

	feedID := insertFeed(t, db)

	// These should be deleted (91 and 120 days old, read).
	old1 := insertOldReadArticle(t, db, feedID, 91)
	old2 := insertOldReadArticle(t, db, feedID, 120)

	// This should be kept (89 days old).
	kept := insertOldReadArticle(t, db, feedID, 89)

	// Set RetentionDays=90 in preferences.
	if _, err := prefStore.Update(ctx, map[string]interface{}{"retention_days": 90}); err != nil {
		t.Fatalf("update prefs: %v", err)
	}

	s := scheduler.NewForTest(feedStore, articleStore, prefStore)
	s.RunRetention(ctx)

	if articleExists(t, db, old1) {
		t.Errorf("article %s (91d old) should have been deleted", old1)
	}
	if articleExists(t, db, old2) {
		t.Errorf("article %s (120d old) should have been deleted", old2)
	}
	if !articleExists(t, db, kept) {
		t.Errorf("article %s (89d old) should have been kept", kept)
	}
}

// TestRetention_DefaultsTo90DaysWhenZero verifies that RetentionDays=0 in
// preferences is treated as 90 days (not "keep forever").
func TestRetention_DefaultsTo90DaysWhenZero(t *testing.T) {
	db := testDB(t)
	feedStore := storage.NewFeedStore(db)
	articleStore := storage.NewArticleStore(db)
	prefStore := storage.NewPreferencesStore(db)
	ctx := context.Background()

	feedID := insertFeed(t, db)
	old := insertOldReadArticle(t, db, feedID, 95)
	recent := insertOldReadArticle(t, db, feedID, 30)

	// RetentionDays=0 in DB row — scheduler should default to 90.
	if _, err := prefStore.Update(ctx, map[string]interface{}{"retention_days": 0}); err != nil {
		t.Fatalf("update prefs: %v", err)
	}

	s := scheduler.NewForTest(feedStore, articleStore, prefStore)
	s.RunRetention(ctx)

	if articleExists(t, db, old) {
		t.Errorf("article %s (95d old, RetentionDays=0→90) should have been deleted", old)
	}
	if !articleExists(t, db, recent) {
		t.Errorf("article %s (30d old) should have been kept", recent)
	}
}

// TestRetention_KeepsStarredAndSaved verifies starred/saved articles are never
// deleted regardless of age.
func TestRetention_KeepsStarredAndSaved(t *testing.T) {
	db := testDB(t)
	feedStore := storage.NewFeedStore(db)
	articleStore := storage.NewArticleStore(db)
	prefStore := storage.NewPreferencesStore(db)
	ctx := context.Background()

	feedID := insertFeed(t, db)

	// Old starred article — must survive.
	starredID := ulid.Make().String()
	createdAt := time.Now().AddDate(0, 0, -200)
	if err := db.Exec(
		`INSERT INTO articles (id, feed_id, guid, guid_feed_id, title, is_read, is_starred, created_at) VALUES (?, ?, ?, ?, ?, true, true, ?)`,
		starredID, feedID, starredID, feedID, "starred", createdAt,
	).Error; err != nil {
		t.Fatalf("insert starred: %v", err)
	}

	// Old saved-for-later article — must survive.
	savedID := ulid.Make().String()
	if err := db.Exec(
		`INSERT INTO articles (id, feed_id, guid, guid_feed_id, title, is_read, is_saved_for_later, created_at) VALUES (?, ?, ?, ?, ?, true, true, ?)`,
		savedID, feedID, savedID, feedID, "saved", createdAt,
	).Error; err != nil {
		t.Fatalf("insert saved: %v", err)
	}

	if _, err := prefStore.Update(ctx, map[string]interface{}{"retention_days": 90}); err != nil {
		t.Fatalf("update prefs: %v", err)
	}

	s := scheduler.NewForTest(feedStore, articleStore, prefStore)
	s.RunRetention(ctx)

	if !articleExists(t, db, starredID) {
		t.Errorf("starred article should have been kept")
	}
	if !articleExists(t, db, savedID) {
		t.Errorf("saved article should have been kept")
	}
}
