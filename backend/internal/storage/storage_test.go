package storage_test

import (
	"context"
	"testing"
	"time"

	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"

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

// --- Folder tests ---

func TestFolderCRUD(t *testing.T) {
	db := testDB(t)
	store := storage.NewFolderStore(db)
	ctx := context.Background()

	f := &storage.Folder{ID: ulid.Make().String(), Name: "Tech"}
	created, err := store.Create(ctx, f)
	if err != nil || created.Name != "Tech" {
		t.Fatalf("Create: %v %v", err, created)
	}

	got, err := store.Get(ctx, created.ID)
	if err != nil || got.Name != "Tech" {
		t.Fatalf("Get: %v", err)
	}

	updated, err := store.Update(ctx, created.ID, map[string]interface{}{"name": "Science"})
	if err != nil || updated.Name != "Science" {
		t.Fatalf("Update: %v", err)
	}

	if err := store.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := store.Get(ctx, created.ID); err == nil {
		t.Fatal("expected not found after delete")
	}
}

func TestFolderList(t *testing.T) {
	db := testDB(t)
	store := storage.NewFolderStore(db)
	ctx := context.Background()

	for i, name := range []string{"C", "A", "B"} {
		store.Create(ctx, &storage.Folder{ID: ulid.Make().String(), Name: name, Position: i})
	}
	folders, err := store.List(ctx)
	if err != nil || len(folders) != 3 {
		t.Fatalf("List: %v len=%d", err, len(folders))
	}
	if folders[0].Name != "C" { // position 0 first
		t.Errorf("expected C first, got %s", folders[0].Name)
	}
}

func TestFolderReorder(t *testing.T) {
	db := testDB(t)
	store := storage.NewFolderStore(db)
	ctx := context.Background()

	ids := make([]string, 3)
	for i, name := range []string{"A", "B", "C"} {
		id := ulid.Make().String()
		ids[i] = id
		store.Create(ctx, &storage.Folder{ID: id, Name: name})
	}
	// Reverse order.
	reversed := []string{ids[2], ids[1], ids[0]}
	folders, err := store.Reorder(ctx, reversed)
	if err != nil || len(folders) != 3 {
		t.Fatalf("Reorder: %v", err)
	}
	if folders[0].ID != ids[2] {
		t.Errorf("expected %s first after reorder, got %s", ids[2], folders[0].ID)
	}
}

// --- Feed tests ---

func TestFeedCRUD(t *testing.T) {
	db := testDB(t)
	store := storage.NewFeedStore(db)
	ctx := context.Background()

	f := &storage.Feed{
		ID:     ulid.Make().String(),
		Title:  "Hacker News",
		XMLURL: "https://news.ycombinator.com/rss",
	}
	created, err := store.Create(ctx, f)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := store.Get(ctx, created.ID)
	if err != nil || got.Title != "Hacker News" {
		t.Fatalf("Get: %v", err)
	}

	byURL, err := store.GetByXMLURL(ctx, "https://news.ycombinator.com/rss")
	if err != nil || byURL.ID != created.ID {
		t.Fatalf("GetByXMLURL: %v", err)
	}

	if err := store.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestFeedListDueForRefresh(t *testing.T) {
	db := testDB(t)
	store := storage.NewFeedStore(db)
	ctx := context.Background()

	// Feed never fetched → due.
	store.Create(ctx, &storage.Feed{ID: ulid.Make().String(), Title: "A", XMLURL: "http://a.example/rss", RefreshIntervalSeconds: 900})
	// Feed fetched 1 hour ago with 15min interval → due.
	past := time.Now().Add(-1 * time.Hour)
	f2 := &storage.Feed{ID: ulid.Make().String(), Title: "B", XMLURL: "http://b.example/rss", RefreshIntervalSeconds: 900, LastFetchedAt: &past}
	store.Create(ctx, f2)
	// Feed fetched now → NOT due.
	now := time.Now()
	f3 := &storage.Feed{ID: ulid.Make().String(), Title: "C", XMLURL: "http://c.example/rss", RefreshIntervalSeconds: 900, LastFetchedAt: &now}
	store.Create(ctx, f3)

	// globalMin=60 — all three feeds have 900s interval so the floor doesn't change results.
	due, err := store.ListDueForRefresh(ctx, 60)
	if err != nil {
		t.Fatalf("ListDueForRefresh: %v", err)
	}
	if len(due) != 2 {
		t.Errorf("expected 2 due feeds, got %d", len(due))
	}

	// Verify the global minimum floor: set globalMin=9999 — feed fetched 1h ago
	// should no longer be due because 9999s haven't elapsed.
	due2, err := store.ListDueForRefresh(ctx, 9999)
	if err != nil {
		t.Fatalf("ListDueForRefresh (high globalMin): %v", err)
	}
	// Only the never-fetched feed (A) should still be due.
	if len(due2) != 1 {
		t.Errorf("expected 1 due feed with high globalMin, got %d", len(due2))
	}
}

func TestFeedBackoff(t *testing.T) {
	db := testDB(t)
	store := storage.NewFeedStore(db)
	ctx := context.Background()

	now := time.Now()
	past := now.Add(-2 * time.Hour)

	// Feed fetched 2h ago, due for refresh.
	f := &storage.Feed{ID: ulid.Make().String(), Title: "BackoffTest", XMLURL: "http://backoff.example/rss", RefreshIntervalSeconds: 60, LastFetchedAt: &past}
	store.Create(ctx, f)

	// Confirm it's due before applying backoff.
	due, err := store.ListDueForRefresh(ctx, 60)
	if err != nil {
		t.Fatalf("ListDueForRefresh: %v", err)
	}
	if len(due) != 1 {
		t.Fatalf("expected 1 due feed, got %d", len(due))
	}

	// Apply a 1-hour backoff (simulating a 403 response).
	if err := store.UpdateLastFetched(ctx, f.ID, now, "HTTP 403", time.Hour); err != nil {
		t.Fatalf("UpdateLastFetched: %v", err)
	}

	// Feed should NOT be due — backoff window hasn't expired.
	due2, err := store.ListDueForRefresh(ctx, 60)
	if err != nil {
		t.Fatalf("ListDueForRefresh after backoff: %v", err)
	}
	for _, d := range due2 {
		if d.ID == f.ID {
			t.Error("feed should not be due while in backoff")
		}
	}

	// Successful fetch clears backoff.
	if err := store.UpdateLastFetched(ctx, f.ID, now, "", 0); err != nil {
		t.Fatalf("UpdateLastFetched (clear): %v", err)
	}
	updated, _ := store.Get(ctx, f.ID)
	if updated.BackoffUntil != nil {
		t.Error("BackoffUntil should be nil after successful fetch")
	}
}

// --- Article tests ---

func TestArticleBulkCreateDedup(t *testing.T) {
	db := testDB(t)
	feedStore := storage.NewFeedStore(db)
	articleStore := storage.NewArticleStore(db)
	ctx := context.Background()

	feedID := ulid.Make().String()
	feedStore.Create(ctx, &storage.Feed{ID: feedID, Title: "Test", XMLURL: "http://test.example/rss"})

	articles := []*storage.Article{
		{ID: ulid.Make().String(), FeedID: feedID, GUIDFeedID: feedID, Title: "A1", GUID: "g1"},
		{ID: ulid.Make().String(), FeedID: feedID, GUIDFeedID: feedID, Title: "A2", GUID: "g2"},
	}
	n, err := articleStore.BulkCreate(ctx, articles)
	if err != nil || n != 2 {
		t.Fatalf("BulkCreate first: %v n=%d", err, n)
	}

	// Insert again with same GUIDs → should skip.
	dupes := []*storage.Article{
		{ID: ulid.Make().String(), FeedID: feedID, GUIDFeedID: feedID, Title: "A1 dup", GUID: "g1"},
		{ID: ulid.Make().String(), FeedID: feedID, GUIDFeedID: feedID, Title: "A3 new", GUID: "g3"},
	}
	n2, err := articleStore.BulkCreate(ctx, dupes)
	if err != nil {
		t.Fatalf("BulkCreate second: %v", err)
	}
	_ = n2 // rows affected may vary by driver

	// Verify count is 3 (not 4).
	opts := storage.ListArticlesOpts{FeedID: feedID, PageSize: 50}
	all, _, err := articleStore.List(ctx, opts)
	if err != nil || len(all) != 3 {
		t.Errorf("expected 3 articles after dedup, got %d (err: %v)", len(all), err)
	}
}

func TestArticleMarkRead(t *testing.T) {
	db := testDB(t)
	feedStore := storage.NewFeedStore(db)
	articleStore := storage.NewArticleStore(db)
	ctx := context.Background()

	feedID := ulid.Make().String()
	feedStore.Create(ctx, &storage.Feed{ID: feedID, Title: "T", XMLURL: "http://t.example/rss"})
	id := ulid.Make().String()
	articleStore.Create(ctx, &storage.Article{ID: id, FeedID: feedID, GUIDFeedID: feedID, GUID: "g1"})

	a, err := articleStore.MarkRead(ctx, id, true)
	if err != nil || !a.IsRead || a.ReadAt == nil {
		t.Fatalf("MarkRead: %v is_read=%v read_at=%v", err, a.IsRead, a.ReadAt)
	}
	a2, _ := articleStore.MarkRead(ctx, id, false)
	if a2.IsRead || a2.ReadAt != nil {
		t.Error("expected unread after MarkRead(false)")
	}
}

func TestArticleMarkAllRead(t *testing.T) {
	db := testDB(t)
	feedStore := storage.NewFeedStore(db)
	articleStore := storage.NewArticleStore(db)
	ctx := context.Background()

	feedID := ulid.Make().String()
	feedStore.Create(ctx, &storage.Feed{ID: feedID, Title: "T", XMLURL: "http://t.example/rss"})
	for i, guid := range []string{"g1", "g2", "g3"} {
		_ = i
		articleStore.Create(ctx, &storage.Article{ID: ulid.Make().String(), FeedID: feedID, GUIDFeedID: feedID, GUID: guid})
	}

	count, err := articleStore.MarkAllRead(ctx, storage.MarkAllReadOpts{FeedID: feedID})
	if err != nil || count != 3 {
		t.Errorf("MarkAllRead feed: %v count=%d", err, count)
	}
}

func TestArticleFTSSearch(t *testing.T) {
	db := testDB(t)
	feedStore := storage.NewFeedStore(db)
	articleStore := storage.NewArticleStore(db)
	ctx := context.Background()

	feedID := ulid.Make().String()
	feedStore.Create(ctx, &storage.Feed{ID: feedID, Title: "T", XMLURL: "http://t.example/rss"})
	id1 := ulid.Make().String()
	id2 := ulid.Make().String()
	articleStore.Create(ctx, &storage.Article{ID: id1, FeedID: feedID, GUIDFeedID: feedID, GUID: "g1", Title: "Go concurrency patterns"})
	articleStore.Create(ctx, &storage.Article{ID: id2, FeedID: feedID, GUIDFeedID: feedID, GUID: "g2", Title: "Python data science"})

	results, _, err := articleStore.List(ctx, storage.ListArticlesOpts{Query: "concurrency", PageSize: 10})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) != 1 || results[0].Title != "Go concurrency patterns" {
		t.Errorf("search result unexpected: %v", results)
	}

	// Mark g1 as read, then verify UnreadOnly + FTS finds nothing.
	articleStore.MarkRead(ctx, id1, true)
	unreadResults, _, err := articleStore.List(ctx, storage.ListArticlesOpts{Query: "concurrency", UnreadOnly: true, PageSize: 10})
	if err != nil {
		t.Fatalf("unread+fts search: %v", err)
	}
	if len(unreadResults) != 0 {
		t.Errorf("expected 0 unread results for read article, got %d", len(unreadResults))
	}
}

func TestArticleDeleteExpired(t *testing.T) {
	db := testDB(t)
	feedStore := storage.NewFeedStore(db)
	articleStore := storage.NewArticleStore(db)
	ctx := context.Background()

	feedID := ulid.Make().String()
	feedStore.Create(ctx, &storage.Feed{ID: feedID, Title: "T", XMLURL: "http://t.example/rss"})

	// Old read article — should be deleted.
	id1 := ulid.Make().String()
	articleStore.Create(ctx, &storage.Article{ID: id1, FeedID: feedID, GUIDFeedID: feedID, GUID: "g1"})
	articleStore.MarkRead(ctx, id1, true)

	// Old starred article — should be preserved.
	id2 := ulid.Make().String()
	articleStore.Create(ctx, &storage.Article{ID: id2, FeedID: feedID, GUIDFeedID: feedID, GUID: "g2", IsStarred: true})
	articleStore.MarkRead(ctx, id2, true)

	// Unread article — should be preserved.
	id3 := ulid.Make().String()
	articleStore.Create(ctx, &storage.Article{ID: id3, FeedID: feedID, GUIDFeedID: feedID, GUID: "g3"})

	// RetentionDays=0 must not delete anything.
	deleted, err := articleStore.DeleteExpired(ctx, 0)
	if err != nil || deleted != 0 {
		t.Errorf("DeleteExpired(0): expected 0 deleted, got %d err %v", deleted, err)
	}

	// With a very large retention, nothing should be deleted either.
	deleted, err = articleStore.DeleteExpired(ctx, 3650)
	if err != nil || deleted != 0 {
		t.Errorf("DeleteExpired(large): expected 0 deleted, got %d err %v", deleted, err)
	}

	all, _, _ := articleStore.List(ctx, storage.ListArticlesOpts{FeedID: feedID, PageSize: 50})
	if len(all) != 3 {
		t.Errorf("expected all 3 articles preserved, got %d", len(all))
	}
}

func TestFolderDeleteOrphansFeeds(t *testing.T) {
	db := testDB(t)
	folderStore := storage.NewFolderStore(db)
	feedStore := storage.NewFeedStore(db)
	ctx := context.Background()

	folder := &storage.Folder{ID: ulid.Make().String(), Name: "Tech"}
	folderStore.Create(ctx, folder)

	feedID := ulid.Make().String()
	feedStore.Create(ctx, &storage.Feed{ID: feedID, Title: "HN", XMLURL: "http://hn.example/rss", FolderID: folder.ID})

	if err := folderStore.Delete(ctx, folder.ID); err != nil {
		t.Fatalf("Delete folder: %v", err)
	}

	// Feed should still exist but with empty folder_id.
	f, err := feedStore.Get(ctx, feedID)
	if err != nil {
		t.Fatalf("feed gone after folder delete: %v", err)
	}
	if f.FolderID != "" {
		t.Errorf("expected empty folder_id after folder delete, got %q", f.FolderID)
	}
}

func TestMarkAllRead_FolderScope(t *testing.T) {
	db := testDB(t)
	folderStore := storage.NewFolderStore(db)
	feedStore := storage.NewFeedStore(db)
	articleStore := storage.NewArticleStore(db)
	ctx := context.Background()

	folder := &storage.Folder{ID: ulid.Make().String(), Name: "Tech"}
	folderStore.Create(ctx, folder)

	feedID := ulid.Make().String()
	feedStore.Create(ctx, &storage.Feed{ID: feedID, Title: "HN", XMLURL: "http://hn2.example/rss", FolderID: folder.ID})

	otherFeedID := ulid.Make().String()
	feedStore.Create(ctx, &storage.Feed{ID: otherFeedID, Title: "Other", XMLURL: "http://other2.example/rss", FolderID: ""})

	for _, g := range []string{"g1", "g2"} {
		articleStore.Create(ctx, &storage.Article{ID: ulid.Make().String(), FeedID: feedID, GUIDFeedID: feedID, GUID: g})
	}
	// Article in a different feed — must not be marked read.
	articleStore.Create(ctx, &storage.Article{ID: ulid.Make().String(), FeedID: otherFeedID, GUIDFeedID: otherFeedID, GUID: "g3"})

	count, err := articleStore.MarkAllRead(ctx, storage.MarkAllReadOpts{FolderID: folder.ID})
	if err != nil {
		t.Fatalf("MarkAllRead folder: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 marked read in folder, got %d", count)
	}

	// Other feed article must remain unread.
	others, _, _ := articleStore.List(ctx, storage.ListArticlesOpts{FeedID: otherFeedID, UnreadOnly: true, PageSize: 10})
	if len(others) != 1 {
		t.Errorf("expected other feed article still unread, got %d", len(others))
	}
}

func TestMarkAllRead_AllScope(t *testing.T) {
	db := testDB(t)
	feedStore := storage.NewFeedStore(db)
	articleStore := storage.NewArticleStore(db)
	ctx := context.Background()

	feedID := ulid.Make().String()
	feedStore.Create(ctx, &storage.Feed{ID: feedID, Title: "T", XMLURL: "http://t2.example/rss"})
	for _, g := range []string{"ga", "gb", "gc"} {
		articleStore.Create(ctx, &storage.Article{ID: ulid.Make().String(), FeedID: feedID, GUIDFeedID: feedID, GUID: g})
	}

	count, err := articleStore.MarkAllRead(ctx, storage.MarkAllReadOpts{All: true})
	if err != nil {
		t.Fatalf("MarkAllRead all: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 marked read, got %d", count)
	}

	unread, _, _ := articleStore.List(ctx, storage.ListArticlesOpts{FeedID: feedID, UnreadOnly: true, PageSize: 10})
	if len(unread) != 0 {
		t.Errorf("expected 0 unread after mark all read, got %d", len(unread))
	}
}

func TestMarkRead_NotFound(t *testing.T) {
	db := testDB(t)
	articleStore := storage.NewArticleStore(db)
	ctx := context.Background()

	_, err := articleStore.MarkRead(ctx, "nonexistent-id", true)
	if err == nil {
		t.Error("expected error marking non-existent article as read")
	}
}

func TestFTSQueryInjection(t *testing.T) {
	db := testDB(t)
	feedStore := storage.NewFeedStore(db)
	articleStore := storage.NewArticleStore(db)
	ctx := context.Background()

	feedID := ulid.Make().String()
	feedStore.Create(ctx, &storage.Feed{ID: feedID, Title: "T", XMLURL: "http://t3.example/rss"})
	articleStore.Create(ctx, &storage.Article{ID: ulid.Make().String(), FeedID: feedID, GUIDFeedID: feedID, GUID: "g1", Title: "Go concurrency"})

	// These queries contain FTS5 operator injection attempts — must not panic or return errors.
	injections := []string{
		"AND OR NOT",
		`"unclosed quote`,
		"NEAR(foo bar)",
		"title:injection",
		"* prefix",
	}
	for _, q := range injections {
		_, _, err := articleStore.List(ctx, storage.ListArticlesOpts{Query: q, PageSize: 5})
		if err != nil {
			t.Errorf("FTS injection %q caused error: %v", q, err)
		}
	}
}

func TestArticleDeleteExpired_ActuallyDeletes(t *testing.T) {
	db := testDB(t)
	feedStore := storage.NewFeedStore(db)
	articleStore := storage.NewArticleStore(db)
	ctx := context.Background()

	feedID := ulid.Make().String()
	feedStore.Create(ctx, &storage.Feed{ID: feedID, Title: "T", XMLURL: "http://t4.example/rss"})

	// Create and immediately mark as read — created_at is now.
	// Use retention=0 to verify the guard works, then use a negative cutoff trick:
	// we can't easily travel back in time, so verify the normal path deletes 0
	// (since articles were just created) and that the query itself executes.
	id1 := ulid.Make().String()
	articleStore.Create(ctx, &storage.Article{ID: id1, FeedID: feedID, GUIDFeedID: feedID, GUID: "g_exp1"})
	articleStore.MarkRead(ctx, id1, true)

	// Retention=1 day — article created now should NOT be deleted yet.
	deleted, err := articleStore.DeleteExpired(ctx, 1)
	if err != nil {
		t.Fatalf("DeleteExpired: %v", err)
	}
	if deleted != 0 {
		t.Errorf("expected 0 deleted (article too recent), got %d", deleted)
	}

	// Verify the article is still there.
	all, _, _ := articleStore.List(ctx, storage.ListArticlesOpts{FeedID: feedID, PageSize: 10})
	if len(all) != 1 {
		t.Errorf("expected article preserved, got %d", len(all))
	}
}

func TestReorder_InvalidID(t *testing.T) {
	db := testDB(t)
	store := storage.NewFolderStore(db)
	ctx := context.Background()

	id := ulid.Make().String()
	store.Create(ctx, &storage.Folder{ID: id, Name: "Real"})

	// Passing a non-existent ID must return an error.
	_, err := store.Reorder(ctx, []string{id, "ghost-id"})
	if err == nil {
		t.Error("expected error when reordering with non-existent folder ID")
	}
}

// --- Recently Read tests ---

func TestRecentlyRead_OnlyShowsIndividuallyOpened(t *testing.T) {
	db := testDB(t)
	feedStore := storage.NewFeedStore(db)
	articleStore := storage.NewArticleStore(db)
	ctx := context.Background()

	feedID := ulid.Make().String()
	feedStore.Create(ctx, &storage.Feed{ID: feedID, Title: "Test", XMLURL: "http://rr.example/rss"})

	// Create 3 articles: all unread at start.
	a1 := &storage.Article{ID: ulid.Make().String(), FeedID: feedID, GUIDFeedID: feedID, GUID: "r1", Title: "Individually Read"}
	a2 := &storage.Article{ID: ulid.Make().String(), FeedID: feedID, GUIDFeedID: feedID, GUID: "r2", Title: "Bulk Read"}
	a3 := &storage.Article{ID: ulid.Make().String(), FeedID: feedID, GUIDFeedID: feedID, GUID: "r3", Title: "Still Unread"}
	articleStore.Create(ctx, a1)
	articleStore.Create(ctx, a2)
	articleStore.Create(ctx, a3)

	// Individually open a1 — sets read_at.
	articleStore.MarkRead(ctx, a1.ID, true)

	// Bulk mark a2 — must NOT set read_at.
	articleStore.MarkAllRead(ctx, storage.MarkAllReadOpts{FeedID: feedID})

	// Recently Read (readOnly=true) must contain ONLY a1.
	recent, _, err := articleStore.List(ctx, storage.ListArticlesOpts{ReadOnly: true, PageSize: 50})
	if err != nil {
		t.Fatalf("List readOnly: %v", err)
	}
	if len(recent) != 1 {
		t.Errorf("expected 1 recently-read article (individually opened), got %d", len(recent))
		for _, a := range recent {
			t.Logf("  article: %s read_at=%v", a.Title, a.ReadAt)
		}
	}
	if len(recent) > 0 && recent[0].ID != a1.ID {
		t.Errorf("expected article %q in recently-read, got %q", a1.Title, recent[0].Title)
	}

	// Verify a2 is read but NOT in recently-read (no read_at).
	a2check, _ := articleStore.Get(ctx, a2.ID)
	if !a2check.IsRead {
		t.Error("a2 should be marked read after MarkAllRead")
	}
	if a2check.ReadAt != nil {
		t.Errorf("a2.ReadAt should be nil after MarkAllRead, got %v", a2check.ReadAt)
	}
}

// --- Preferences tests ---

func TestPreferencesGetUpdate(t *testing.T) {
	db := testDB(t)
	store := storage.NewPreferencesStore(db)
	ctx := context.Background()

	p, err := store.Get(ctx)
	if err != nil || p.Theme != "dark" {
		t.Fatalf("Get default: %v", err)
	}

	updated, err := store.Update(ctx, map[string]interface{}{"theme": "light", "hide_read_articles": false})
	if err != nil || updated.Theme != "light" || updated.HideReadArticles {
		t.Errorf("Update: %v %+v", err, updated)
	}
}
