package service_test

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"

	pb "github.com/plexreader/plexreader/backend/gen/plexreader/v1"
	"github.com/plexreader/plexreader/backend/internal/service"
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

func mustCreateFeed(t *testing.T, store storage.FeedStore, xmlURL string) string {
	t.Helper()
	id := ulid.Make().String()
	_, err := store.Create(context.Background(), &storage.Feed{
		ID:                     id,
		Title:                  "Feed",
		XMLURL:                 xmlURL,
		RefreshIntervalSeconds: 900,
	})
	if err != nil {
		t.Fatalf("create feed: %v", err)
	}
	return id
}

func mustCreateArticle(t *testing.T, store storage.ArticleStore, feedID, guid string) string {
	t.Helper()
	id := ulid.Make().String()
	_, err := store.Create(context.Background(), &storage.Article{
		ID:         id,
		FeedID:     feedID,
		GUIDFeedID: feedID,
		GUID:       guid,
		Title:      "Article " + guid,
	})
	if err != nil {
		t.Fatalf("create article: %v", err)
	}
	return id
}

// --- ArticleService ---

func TestMarkAllAsRead_NoScopeNoIDs_ReturnsInvalidArgument(t *testing.T) {
	db := testDB(t)
	svc := service.NewArticleService(storage.NewArticleStore(db))

	_, err := svc.MarkAllAsRead(context.Background(), connect.NewRequest(&pb.MarkAllAsReadRequest{}))
	if err == nil {
		t.Fatal("expected error when scope and article_ids both empty")
	}
	ce, ok := err.(*connect.Error)
	if !ok || ce.Code() != connect.CodeInvalidArgument {
		t.Errorf("expected CodeInvalidArgument, got %v", err)
	}
}

func TestMarkAllAsRead_ArticleIDs_OnlyMarksThose(t *testing.T) {
	db := testDB(t)
	feedStore := storage.NewFeedStore(db)
	articleStore := storage.NewArticleStore(db)
	svc := service.NewArticleService(articleStore)
	ctx := context.Background()

	fid := mustCreateFeed(t, feedStore, "http://test.example/rss")
	id1 := mustCreateArticle(t, articleStore, fid, "g1")
	id2 := mustCreateArticle(t, articleStore, fid, "g2")
	id3 := mustCreateArticle(t, articleStore, fid, "g3")

	resp, err := svc.MarkAllAsRead(ctx, connect.NewRequest(&pb.MarkAllAsReadRequest{
		ArticleIds: []string{id1, id2},
	}))
	if err != nil {
		t.Fatalf("MarkAllAsRead: %v", err)
	}
	if resp.Msg.Count != 2 {
		t.Errorf("expected 2 marked, got %d", resp.Msg.Count)
	}

	a3, _ := articleStore.Get(ctx, id3)
	if a3.IsRead {
		t.Error("id3 should remain unread")
	}
}

func TestMarkAllAsRead_FeedScope_OnlyMarksFeedArticles(t *testing.T) {
	db := testDB(t)
	feedStore := storage.NewFeedStore(db)
	articleStore := storage.NewArticleStore(db)
	svc := service.NewArticleService(articleStore)
	ctx := context.Background()

	fid1 := mustCreateFeed(t, feedStore, "http://feed1.example/rss")
	fid2 := mustCreateFeed(t, feedStore, "http://feed2.example/rss")
	mustCreateArticle(t, articleStore, fid1, "g1")
	mustCreateArticle(t, articleStore, fid1, "g2")
	mustCreateArticle(t, articleStore, fid2, "g3")

	resp, err := svc.MarkAllAsRead(ctx, connect.NewRequest(&pb.MarkAllAsReadRequest{
		Scope: &pb.MarkAllAsReadRequest_FeedId{FeedId: fid1},
	}))
	if err != nil {
		t.Fatalf("MarkAllAsRead: %v", err)
	}
	if resp.Msg.Count != 2 {
		t.Errorf("expected 2, got %d", resp.Msg.Count)
	}

	unread, _, _ := articleStore.List(ctx, storage.ListArticlesOpts{
		FeedID: fid2, UnreadOnly: true, PageSize: 10,
	})
	if len(unread) != 1 {
		t.Errorf("fid2 article should still be unread, got %d unread", len(unread))
	}
}

func TestListArticles_UnreadOnlyDefault(t *testing.T) {
	db := testDB(t)
	feedStore := storage.NewFeedStore(db)
	articleStore := storage.NewArticleStore(db)
	svc := service.NewArticleService(articleStore)
	ctx := context.Background()

	fid := mustCreateFeed(t, feedStore, "http://unread.example/rss")
	id1 := mustCreateArticle(t, articleStore, fid, "read")
	mustCreateArticle(t, articleStore, fid, "unread")
	articleStore.MarkRead(ctx, id1, true)

	resp, err := svc.ListArticles(ctx, connect.NewRequest(&pb.ListArticlesRequest{
		FeedId:     fid,
		UnreadOnly: true,
		PageSize:   10,
	}))
	if err != nil {
		t.Fatalf("ListArticles: %v", err)
	}
	if len(resp.Msg.Articles) != 1 {
		t.Errorf("expected 1 unread article, got %d", len(resp.Msg.Articles))
	}
}

// --- FeedService ---

func TestCreateFeed_NilBody_ReturnsInvalidArgument(t *testing.T) {
	db := testDB(t)
	svc := service.NewFeedService(storage.NewFeedStore(db), storage.NewFolderStore(db), nil)

	_, err := svc.CreateFeed(context.Background(), connect.NewRequest(&pb.CreateFeedRequest{}))
	if err == nil {
		t.Fatal("expected error for nil feed body")
	}
	ce, ok := err.(*connect.Error)
	if !ok || ce.Code() != connect.CodeInvalidArgument {
		t.Errorf("expected CodeInvalidArgument, got %v", err)
	}
}

func TestImportOPML_DeduplicatesOnReimport(t *testing.T) {
	db := testDB(t)
	feedStore := storage.NewFeedStore(db)
	folderStore := storage.NewFolderStore(db)
	svc := service.NewFeedService(feedStore, folderStore, nil)
	ctx := context.Background()

	opml := []byte(`<?xml version="1.0"?>
<opml version="2.0">
  <head><title>Test</title></head>
  <body>
    <outline text="Go Blog" type="rss" xmlUrl="https://go.dev/blog/feed.atom"/>
  </body>
</opml>`)

	// First import — should create 1 feed.
	resp1, err := svc.ImportOPML(ctx, connect.NewRequest(&pb.ImportOPMLRequest{OpmlContent: opml}))
	if err != nil {
		t.Fatalf("first import: %v", err)
	}
	if resp1.Msg.FeedsCreated != 1 {
		t.Errorf("expected 1 feed, got %d", resp1.Msg.FeedsCreated)
	}

	// Second import — must skip existing.
	resp2, err := svc.ImportOPML(ctx, connect.NewRequest(&pb.ImportOPMLRequest{OpmlContent: opml}))
	if err != nil {
		t.Fatalf("second import: %v", err)
	}
	if resp2.Msg.FeedsCreated != 0 {
		t.Errorf("expected 0 new feeds on re-import, got %d", resp2.Msg.FeedsCreated)
	}

	// Exactly 1 feed in DB.
	feeds, _, _ := feedStore.List(ctx, storage.ListFeedsOpts{PageSize: 50})
	if len(feeds) != 1 {
		t.Errorf("expected 1 feed in db, got %d", len(feeds))
	}
}

func TestImportOPML_CreatesFoldersAndAssociatesFeeds(t *testing.T) {
	db := testDB(t)
	feedStore := storage.NewFeedStore(db)
	folderStore := storage.NewFolderStore(db)
	svc := service.NewFeedService(feedStore, folderStore, nil)
	ctx := context.Background()

	opml := []byte(`<?xml version="1.0"?>
<opml version="2.0">
  <head><title>Test</title></head>
  <body>
    <outline text="Technology">
      <outline text="Go Blog" type="rss" xmlUrl="https://go.dev/blog/feed.atom"/>
      <outline text="HN" type="rss" xmlUrl="https://news.ycombinator.com/rss"/>
    </outline>
    <outline text="Uncategorized" type="rss" xmlUrl="https://example.com/rss"/>
  </body>
</opml>`)

	resp, err := svc.ImportOPML(ctx, connect.NewRequest(&pb.ImportOPMLRequest{OpmlContent: opml}))
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if resp.Msg.FoldersCreated != 1 {
		t.Errorf("expected 1 folder, got %d", resp.Msg.FoldersCreated)
	}
	if resp.Msg.FeedsCreated != 3 {
		t.Errorf("expected 3 feeds, got %d", resp.Msg.FeedsCreated)
	}

	folders, _ := folderStore.List(ctx)
	if len(folders) != 1 || folders[0].Name != "Technology" {
		t.Errorf("Technology folder missing: %v", folders)
	}

	techFeeds, _, _ := feedStore.List(ctx, storage.ListFeedsOpts{FolderID: folders[0].ID, PageSize: 10})
	if len(techFeeds) != 2 {
		t.Errorf("expected 2 feeds in Technology, got %d", len(techFeeds))
	}
}

func TestImportOPML_InvalidXML_ReturnsInvalidArgument(t *testing.T) {
	db := testDB(t)
	svc := service.NewFeedService(storage.NewFeedStore(db), storage.NewFolderStore(db), nil)

	_, err := svc.ImportOPML(context.Background(), connect.NewRequest(&pb.ImportOPMLRequest{
		OpmlContent: []byte("not xml"),
	}))
	if err == nil {
		t.Fatal("expected error for invalid OPML")
	}
	ce, ok := err.(*connect.Error)
	if !ok || ce.Code() != connect.CodeInvalidArgument {
		t.Errorf("expected CodeInvalidArgument, got %v", err)
	}
}

func TestImportOPML_EmptyContent_ReturnsInvalidArgument(t *testing.T) {
	db := testDB(t)
	svc := service.NewFeedService(storage.NewFeedStore(db), storage.NewFolderStore(db), nil)

	_, err := svc.ImportOPML(context.Background(), connect.NewRequest(&pb.ImportOPMLRequest{
		OpmlContent: []byte(""),
	}))
	if err == nil {
		t.Fatal("expected error for empty content")
	}
	ce, ok := err.(*connect.Error)
	if !ok || ce.Code() != connect.CodeInvalidArgument {
		t.Errorf("expected CodeInvalidArgument, got %v", err)
	}
}

func TestImportOPML_TooLarge_ReturnsInvalidArgument(t *testing.T) {
	db := testDB(t)
	svc := service.NewFeedService(storage.NewFeedStore(db), storage.NewFolderStore(db), nil)

	big := make([]byte, 1<<20+1) // 1 MiB + 1 byte
	_, err := svc.ImportOPML(context.Background(), connect.NewRequest(&pb.ImportOPMLRequest{
		OpmlContent: big,
	}))
	if err == nil {
		t.Fatal("expected error for oversized content")
	}
	ce, ok := err.(*connect.Error)
	if !ok || ce.Code() != connect.CodeInvalidArgument {
		t.Errorf("expected CodeInvalidArgument, got %v", err)
	}
}

func TestImportOPML_UncategorizedFeeds(t *testing.T) {
	db := testDB(t)
	feedStore := storage.NewFeedStore(db)
	folderStore := storage.NewFolderStore(db)
	svc := service.NewFeedService(feedStore, folderStore, nil)
	ctx := context.Background()

	opml := []byte(`<?xml version="1.0"?>
<opml version="2.0">
  <head><title>Test</title></head>
  <body>
    <outline text="HN" type="rss" xmlUrl="https://news.ycombinator.com/rss"/>
    <outline text="TechCrunch" type="rss" xmlUrl="https://techcrunch.com/feed/"/>
  </body>
</opml>`)

	resp, err := svc.ImportOPML(ctx, connect.NewRequest(&pb.ImportOPMLRequest{OpmlContent: opml}))
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if resp.Msg.FeedsCreated != 2 {
		t.Errorf("expected 2 feeds, got %d", resp.Msg.FeedsCreated)
	}
	if resp.Msg.FoldersCreated != 0 {
		t.Errorf("expected 0 folders, got %d", resp.Msg.FoldersCreated)
	}

	// Verify feeds have no folder assignment.
	feeds, _, err := feedStore.List(ctx, storage.ListFeedsOpts{PageSize: 50})
	if err != nil {
		t.Fatalf("list feeds: %v", err)
	}
	if len(feeds) != 2 {
		t.Fatalf("expected 2 feeds in db, got %d", len(feeds))
	}
	for _, f := range feeds {
		if f.FolderID != "" {
			t.Errorf("uncategorized feed %q should have no folder, got %q", f.XMLURL, f.FolderID)
		}
	}
}

func TestImportOPML_ExistingFolderNotDuplicated(t *testing.T) {
	db := testDB(t)
	feedStore := storage.NewFeedStore(db)
	folderStore := storage.NewFolderStore(db)
	svc := service.NewFeedService(feedStore, folderStore, nil)
	ctx := context.Background()

	// Pre-create the Technology folder.
	_, err := folderStore.Create(ctx, &storage.Folder{
		ID:   "01JEXISTINGFOLDER000000000",
		Name: "Technology",
	})
	if err != nil {
		t.Fatalf("pre-create folder: %v", err)
	}

	opml := []byte(`<?xml version="1.0"?>
<opml version="2.0">
  <head><title>Test</title></head>
  <body>
    <outline text="Technology">
      <outline text="HN" type="rss" xmlUrl="https://news.ycombinator.com/rss"/>
    </outline>
  </body>
</opml>`)

	resp, err := svc.ImportOPML(ctx, connect.NewRequest(&pb.ImportOPMLRequest{OpmlContent: opml}))
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if resp.Msg.FoldersCreated != 0 {
		t.Errorf("expected 0 new folders (already exists), got %d", resp.Msg.FoldersCreated)
	}
	if resp.Msg.FeedsCreated != 1 {
		t.Errorf("expected 1 feed, got %d", resp.Msg.FeedsCreated)
	}

	folders, err := folderStore.List(ctx)
	if err != nil {
		t.Fatalf("list folders: %v", err)
	}
	if len(folders) != 1 {
		t.Errorf("expected exactly 1 folder, got %d", len(folders))
	}

	// Feed must be in the pre-existing folder.
	feeds, _, _ := feedStore.List(ctx, storage.ListFeedsOpts{FolderID: "01JEXISTINGFOLDER000000000", PageSize: 10})
	if len(feeds) != 1 {
		t.Errorf("expected feed in pre-existing folder, got %d feeds there", len(feeds))
	}
}

func TestImportOPML_PartialSuccess_MissingXMLURL(t *testing.T) {
	db := testDB(t)
	feedStore := storage.NewFeedStore(db)
	folderStore := storage.NewFolderStore(db)
	svc := service.NewFeedService(feedStore, folderStore, nil)
	ctx := context.Background()

	// One feed has xmlUrl, one is a folder outline with no xmlUrl and no children feeds with xmlUrl.
	opml := []byte(`<?xml version="1.0"?>
<opml version="2.0">
  <head><title>Test</title></head>
  <body>
    <outline text="HN" type="rss" xmlUrl="https://news.ycombinator.com/rss"/>
    <outline text="bad" type="rss" xmlUrl=""/>
  </body>
</opml>`)

	resp, err := svc.ImportOPML(ctx, connect.NewRequest(&pb.ImportOPMLRequest{OpmlContent: opml}))
	if err != nil {
		t.Fatalf("import should partially succeed: %v", err)
	}
	// HN should be created, the empty-URL one skipped.
	if resp.Msg.FeedsCreated != 1 {
		t.Errorf("expected 1 feed created, got %d", resp.Msg.FeedsCreated)
	}
	if resp.Msg.FeedsSkipped != 1 {
		t.Errorf("expected 1 feed skipped, got %d", resp.Msg.FeedsSkipped)
	}
	if len(resp.Msg.Errors) != 1 {
		t.Errorf("expected 1 error entry, got %d", len(resp.Msg.Errors))
	}

	feeds, _, _ := feedStore.List(ctx, storage.ListFeedsOpts{PageSize: 10})
	if len(feeds) != 1 {
		t.Errorf("expected 1 feed in db, got %d", len(feeds))
	}
}

func TestImportOPML_MultipleFoldersMultipleFeeds(t *testing.T) {
	db := testDB(t)
	feedStore := storage.NewFeedStore(db)
	folderStore := storage.NewFolderStore(db)
	svc := service.NewFeedService(feedStore, folderStore, nil)
	ctx := context.Background()

	opml := []byte(`<?xml version="1.0"?>
<opml version="1.0">
  <head><title>PlexReader Sample Feeds</title></head>
  <body>
    <outline text="Technology" title="Technology">
      <outline text="Hacker News" title="Hacker News" type="rss"
        xmlUrl="https://news.ycombinator.com/rss" htmlUrl="https://news.ycombinator.com/"/>
      <outline text="TechCrunch" title="TechCrunch" type="rss"
        xmlUrl="https://techcrunch.com/feed/" htmlUrl="https://techcrunch.com/"/>
      <outline text="The Verge" title="The Verge" type="rss"
        xmlUrl="https://www.theverge.com/rss/index.xml" htmlUrl="https://www.theverge.com/"/>
    </outline>
    <outline text="Science" title="Science">
      <outline text="NASA" title="NASA" type="rss"
        xmlUrl="https://www.nasa.gov/rss/dyn/breaking_news.rss" htmlUrl="https://www.nasa.gov/"/>
    </outline>
    <outline text="Uncategorized" type="rss" xmlUrl="https://go.dev/blog/feed.atom"/>
  </body>
</opml>`)

	resp, err := svc.ImportOPML(ctx, connect.NewRequest(&pb.ImportOPMLRequest{OpmlContent: opml}))
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if resp.Msg.FoldersCreated != 2 {
		t.Errorf("expected 2 folders, got %d", resp.Msg.FoldersCreated)
	}
	if resp.Msg.FeedsCreated != 5 {
		t.Errorf("expected 5 feeds, got %d", resp.Msg.FeedsCreated)
	}
	if len(resp.Msg.Errors) != 0 {
		t.Errorf("expected no errors, got: %v", resp.Msg.Errors)
	}

	folders, _ := folderStore.List(ctx)
	if len(folders) != 2 {
		t.Errorf("expected 2 folders in db, got %d", len(folders))
	}

	folderByName := make(map[string]string)
	for _, f := range folders {
		folderByName[f.Name] = f.ID
	}

	techID, ok := folderByName["Technology"]
	if !ok {
		t.Fatal("Technology folder missing")
	}
	techFeeds, _, _ := feedStore.List(ctx, storage.ListFeedsOpts{FolderID: techID, PageSize: 10})
	if len(techFeeds) != 3 {
		t.Errorf("expected 3 feeds in Technology, got %d", len(techFeeds))
	}

	sciID, ok := folderByName["Science"]
	if !ok {
		t.Fatal("Science folder missing")
	}
	sciFeeds, _, _ := feedStore.List(ctx, storage.ListFeedsOpts{FolderID: sciID, PageSize: 10})
	if len(sciFeeds) != 1 {
		t.Errorf("expected 1 feed in Science, got %d", len(sciFeeds))
	}

	// One root-level uncategorized feed.
	allFeeds, _, _ := feedStore.List(ctx, storage.ListFeedsOpts{PageSize: 50})
	var uncategorized int
	for _, f := range allFeeds {
		if f.FolderID == "" {
			uncategorized++
		}
	}
	if uncategorized != 1 {
		t.Errorf("expected 1 uncategorized feed, got %d", uncategorized)
	}
}

func TestExportOPML_RoundTrip(t *testing.T) {
	db := testDB(t)
	feedStore := storage.NewFeedStore(db)
	folderStore := storage.NewFolderStore(db)
	svc := service.NewFeedService(feedStore, folderStore, nil)
	ctx := context.Background()

	opml := []byte(`<?xml version="1.0"?>
<opml version="1.0">
  <head><title>Export Test</title></head>
  <body>
    <outline text="Tech" title="Tech">
      <outline text="HN" type="rss" xmlUrl="https://news.ycombinator.com/rss" htmlUrl="https://news.ycombinator.com/"/>
      <outline text="TC" type="rss" xmlUrl="https://techcrunch.com/feed/" htmlUrl="https://techcrunch.com/"/>
    </outline>
  </body>
</opml>`)

	// Import first.
	_, err := svc.ImportOPML(ctx, connect.NewRequest(&pb.ImportOPMLRequest{OpmlContent: opml}))
	if err != nil {
		t.Fatalf("import: %v", err)
	}

	// Export.
	exportResp, err := svc.ExportOPML(ctx, connect.NewRequest(&pb.ExportOPMLRequest{}))
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	if len(exportResp.Msg.OpmlContent) == 0 {
		t.Fatal("export returned empty content")
	}

	// Re-import the exported content — should deduplicate, 0 new feeds.
	resp2, err := svc.ImportOPML(ctx, connect.NewRequest(&pb.ImportOPMLRequest{
		OpmlContent: exportResp.Msg.OpmlContent,
	}))
	if err != nil {
		t.Fatalf("re-import: %v", err)
	}
	if resp2.Msg.FeedsCreated != 0 {
		t.Errorf("re-import should create 0 feeds (all duplicates), got %d", resp2.Msg.FeedsCreated)
	}

	// Still exactly 2 feeds.
	feeds, _, _ := feedStore.List(ctx, storage.ListFeedsOpts{PageSize: 50})
	if len(feeds) != 2 {
		t.Errorf("expected 2 feeds after round-trip, got %d", len(feeds))
	}
}

// ── Article action integration tests ─────────────────────────────────────────

func TestStarArticle_TogglesStarred(t *testing.T) {
	db := testDB(t)
	feedStore := storage.NewFeedStore(db)
	articleStore := storage.NewArticleStore(db)
	svc := service.NewArticleService(articleStore)
	ctx := context.Background()

	fid := mustCreateFeed(t, feedStore, "http://star.example/rss")
	aid := mustCreateArticle(t, articleStore, fid, "g1")

	// Star it.
	resp, err := svc.StarArticle(ctx, connect.NewRequest(&pb.StarArticleRequest{Id: aid, Starred: true}))
	if err != nil {
		t.Fatalf("StarArticle: %v", err)
	}
	if !resp.Msg.IsStarred {
		t.Error("expected is_starred=true after starring")
	}

	// Unstar it.
	resp2, err := svc.StarArticle(ctx, connect.NewRequest(&pb.StarArticleRequest{Id: aid, Starred: false}))
	if err != nil {
		t.Fatalf("StarArticle unstar: %v", err)
	}
	if resp2.Msg.IsStarred {
		t.Error("expected is_starred=false after unstarring")
	}

	// Verify persisted.
	a, _ := articleStore.Get(ctx, aid)
	if a.IsStarred {
		t.Error("expected IsStarred=false in DB after unstar")
	}
}

func TestSaveForLater_TogglesFlag(t *testing.T) {
	db := testDB(t)
	feedStore := storage.NewFeedStore(db)
	articleStore := storage.NewArticleStore(db)
	svc := service.NewArticleService(articleStore)
	ctx := context.Background()

	fid := mustCreateFeed(t, feedStore, "http://save.example/rss")
	aid := mustCreateArticle(t, articleStore, fid, "g1")

	// Save for later.
	resp, err := svc.SaveForLater(ctx, connect.NewRequest(&pb.SaveForLaterRequest{Id: aid, Saved: true}))
	if err != nil {
		t.Fatalf("SaveForLater: %v", err)
	}
	if !resp.Msg.IsSavedForLater {
		t.Error("expected is_saved_for_later=true")
	}

	// Unsave.
	resp2, err := svc.SaveForLater(ctx, connect.NewRequest(&pb.SaveForLaterRequest{Id: aid, Saved: false}))
	if err != nil {
		t.Fatalf("SaveForLater unsave: %v", err)
	}
	if resp2.Msg.IsSavedForLater {
		t.Error("expected is_saved_for_later=false after unsave")
	}

	// Verify persisted.
	a, _ := articleStore.Get(ctx, aid)
	if a.IsSavedForLater {
		t.Error("expected IsSavedForLater=false in DB after unsave")
	}
}

func TestMarkAsRead_ThenMarkAsUnread(t *testing.T) {
	db := testDB(t)
	feedStore := storage.NewFeedStore(db)
	articleStore := storage.NewArticleStore(db)
	svc := service.NewArticleService(articleStore)
	ctx := context.Background()

	fid := mustCreateFeed(t, feedStore, "http://readunread.example/rss")
	aid := mustCreateArticle(t, articleStore, fid, "g1")

	// Mark read.
	resp, err := svc.MarkAsRead(ctx, connect.NewRequest(&pb.MarkAsReadRequest{Id: aid}))
	if err != nil {
		t.Fatalf("MarkAsRead: %v", err)
	}
	if !resp.Msg.IsRead {
		t.Error("expected is_read=true")
	}

	// Mark unread.
	resp2, err := svc.MarkAsUnread(ctx, connect.NewRequest(&pb.MarkAsUnreadRequest{Id: aid}))
	if err != nil {
		t.Fatalf("MarkAsUnread: %v", err)
	}
	if resp2.Msg.IsRead {
		t.Error("expected is_read=false after mark unread")
	}

	// Verify persisted.
	a, _ := articleStore.Get(ctx, aid)
	if a.IsRead {
		t.Error("expected IsRead=false in DB after MarkAsUnread")
	}
}

func TestMarkAllAsRead_FolderScope_OnlyMarksFolderArticles(t *testing.T) {
	db := testDB(t)
	feedStore := storage.NewFeedStore(db)
	folderStore := storage.NewFolderStore(db)
	articleStore := storage.NewArticleStore(db)
	svc := service.NewArticleService(articleStore)
	ctx := context.Background()

	// Two folders, feeds in each.
	folderA, _ := folderStore.Create(ctx, &storage.Folder{ID: ulid.Make().String(), Name: "FolderA"})
	folderB, _ := folderStore.Create(ctx, &storage.Folder{ID: ulid.Make().String(), Name: "FolderB"})
	fidA := ulid.Make().String()
	feedStore.Create(ctx, &storage.Feed{ID: fidA, Title: "FeedA", XMLURL: "http://a.example/rss", FolderID: folderA.ID, RefreshIntervalSeconds: 900})
	fidB := ulid.Make().String()
	feedStore.Create(ctx, &storage.Feed{ID: fidB, Title: "FeedB", XMLURL: "http://b.example/rss", FolderID: folderB.ID, RefreshIntervalSeconds: 900})

	mustCreateArticle(t, articleStore, fidA, "a1")
	mustCreateArticle(t, articleStore, fidA, "a2")
	bID := mustCreateArticle(t, articleStore, fidB, "b1")

	resp, err := svc.MarkAllAsRead(ctx, connect.NewRequest(&pb.MarkAllAsReadRequest{
		Scope: &pb.MarkAllAsReadRequest_FolderId{FolderId: folderA.ID},
	}))
	if err != nil {
		t.Fatalf("MarkAllAsRead folder: %v", err)
	}
	if resp.Msg.Count != 2 {
		t.Errorf("expected 2 marked, got %d", resp.Msg.Count)
	}

	// FolderB article must remain unread.
	b, _ := articleStore.Get(ctx, bID)
	if b.IsRead {
		t.Error("FolderB article should NOT be marked read")
	}
}

func TestMarkAllAsRead_AllScope_MarksEverything(t *testing.T) {
	db := testDB(t)
	feedStore := storage.NewFeedStore(db)
	articleStore := storage.NewArticleStore(db)
	svc := service.NewArticleService(articleStore)
	ctx := context.Background()

	fid1 := mustCreateFeed(t, feedStore, "http://all1.example/rss")
	fid2 := mustCreateFeed(t, feedStore, "http://all2.example/rss")
	mustCreateArticle(t, articleStore, fid1, "g1")
	mustCreateArticle(t, articleStore, fid1, "g2")
	mustCreateArticle(t, articleStore, fid2, "g3")

	resp, err := svc.MarkAllAsRead(ctx, connect.NewRequest(&pb.MarkAllAsReadRequest{
		Scope: &pb.MarkAllAsReadRequest_All{All: true},
	}))
	if err != nil {
		t.Fatalf("MarkAllAsRead all: %v", err)
	}
	if resp.Msg.Count != 3 {
		t.Errorf("expected 3 marked, got %d", resp.Msg.Count)
	}

	// All articles should now be read.
	unread, _, _ := articleStore.List(ctx, storage.ListArticlesOpts{UnreadOnly: true, PageSize: 50})
	if len(unread) != 0 {
		t.Errorf("expected 0 unread after mark-all, got %d", len(unread))
	}
}

func TestStarArticle_NotFound_ReturnsError(t *testing.T) {
	db := testDB(t)
	svc := service.NewArticleService(storage.NewArticleStore(db))

	_, err := svc.StarArticle(context.Background(), connect.NewRequest(&pb.StarArticleRequest{
		Id: "nonexistent-id", Starred: true,
	}))
	if err == nil {
		t.Fatal("expected error for nonexistent article")
	}
}

func TestSaveForLater_NotFound_ReturnsError(t *testing.T) {
	db := testDB(t)
	svc := service.NewArticleService(storage.NewArticleStore(db))

	_, err := svc.SaveForLater(context.Background(), connect.NewRequest(&pb.SaveForLaterRequest{
		Id: "nonexistent-id", Saved: true,
	}))
	if err == nil {
		t.Fatal("expected error for nonexistent article")
	}
}

// --- FolderService ---

func mustCreateFolder(t *testing.T, store storage.FolderStore, name string) string {
	t.Helper()
	id := ulid.Make().String()
	_, err := store.Create(context.Background(), &storage.Folder{ID: id, Name: name})
	if err != nil {
		t.Fatalf("create folder: %v", err)
	}
	return id
}

func TestDeleteFolder_RemovesFromList(t *testing.T) {
	db := testDB(t)
	folderStore := storage.NewFolderStore(db)
	svc := service.NewFolderService(folderStore)
	ctx := context.Background()

	fid := mustCreateFolder(t, folderStore, "Tech")

	_, err := svc.DeleteFolder(ctx, connect.NewRequest(&pb.DeleteFolderRequest{Id: fid}))
	if err != nil {
		t.Fatalf("DeleteFolder: %v", err)
	}

	resp, err := svc.ListFolders(ctx, connect.NewRequest(&pb.ListFoldersRequest{}))
	if err != nil {
		t.Fatalf("ListFolders: %v", err)
	}
	for _, f := range resp.Msg.Folders {
		if f.Id == fid {
			t.Errorf("deleted folder %s still appears in ListFolders", fid)
		}
	}
}

func TestDeleteFolder_Idempotent_NoErrorForMissing(t *testing.T) {
	db := testDB(t)
	svc := service.NewFolderService(storage.NewFolderStore(db))

	// Deleting a nonexistent folder is idempotent — no error returned.
	_, err := svc.DeleteFolder(context.Background(), connect.NewRequest(&pb.DeleteFolderRequest{Id: "nonexistent"}))
	if err != nil {
		t.Fatalf("expected nil for nonexistent folder, got %v", err)
	}
}

func TestUpdateFolder_Rename_ReflectedInList(t *testing.T) {
	db := testDB(t)
	folderStore := storage.NewFolderStore(db)
	svc := service.NewFolderService(folderStore)
	ctx := context.Background()

	fid := mustCreateFolder(t, folderStore, "OldName")

	_, err := svc.UpdateFolder(ctx, connect.NewRequest(&pb.UpdateFolderRequest{
		Folder: &pb.Folder{Id: fid, Name: "NewName"},
	}))
	if err != nil {
		t.Fatalf("UpdateFolder: %v", err)
	}

	resp, err := svc.ListFolders(ctx, connect.NewRequest(&pb.ListFoldersRequest{}))
	if err != nil {
		t.Fatalf("ListFolders: %v", err)
	}
	found := false
	for _, f := range resp.Msg.Folders {
		if f.Id == fid {
			found = true
			if f.Name != "NewName" {
				t.Errorf("expected name NewName, got %s", f.Name)
			}
		}
	}
	if !found {
		t.Errorf("updated folder %s not found in ListFolders", fid)
	}
}

// --- FeedService delete ---

func TestDeleteFeed_RemovesFromList(t *testing.T) {
	db := testDB(t)
	feedStore := storage.NewFeedStore(db)
	folderStore := storage.NewFolderStore(db)
	svc := service.NewFeedService(feedStore, folderStore, nil)
	ctx := context.Background()

	fid := mustCreateFeed(t, feedStore, "http://example.com/rss")

	_, err := svc.DeleteFeed(ctx, connect.NewRequest(&pb.DeleteFeedRequest{Id: fid}))
	if err != nil {
		t.Fatalf("DeleteFeed: %v", err)
	}

	resp, err := svc.ListFeeds(ctx, connect.NewRequest(&pb.ListFeedsRequest{}))
	if err != nil {
		t.Fatalf("ListFeeds: %v", err)
	}
	for _, f := range resp.Msg.Feeds {
		if f.Id == fid {
			t.Errorf("deleted feed %s still appears in ListFeeds", fid)
		}
	}
}

func TestDeleteFeed_Idempotent_NoErrorForMissing(t *testing.T) {
	db := testDB(t)
	svc := service.NewFeedService(storage.NewFeedStore(db), storage.NewFolderStore(db), nil)

	// Deleting a nonexistent feed is idempotent — no error returned.
	_, err := svc.DeleteFeed(context.Background(), connect.NewRequest(&pb.DeleteFeedRequest{Id: "nonexistent"}))
	if err != nil {
		t.Fatalf("expected nil for nonexistent feed, got %v", err)
	}
}

// --- MarkAllAsRead from sidebar (feed scope, folder scope, all) ---

func TestMarkAllAsRead_FeedScope_UnreadCountDropsToZero(t *testing.T) {
	db := testDB(t)
	feedStore := storage.NewFeedStore(db)
	articleStore := storage.NewArticleStore(db)
	svc := service.NewArticleService(articleStore)
	ctx := context.Background()

	fid := mustCreateFeed(t, feedStore, "http://sidebar-feed.example/rss")
	mustCreateArticle(t, articleStore, fid, "s1")
	mustCreateArticle(t, articleStore, fid, "s2")
	mustCreateArticle(t, articleStore, fid, "s3")

	resp, err := svc.MarkAllAsRead(ctx, connect.NewRequest(&pb.MarkAllAsReadRequest{
		Scope: &pb.MarkAllAsReadRequest_FeedId{FeedId: fid},
	}))
	if err != nil {
		t.Fatalf("MarkAllAsRead by feed: %v", err)
	}
	if resp.Msg.Count != 3 {
		t.Errorf("expected 3 marked, got %d", resp.Msg.Count)
	}

	// Verify unread count is now zero for this feed.
	listResp, err := svc.ListArticles(ctx, connect.NewRequest(&pb.ListArticlesRequest{
		FeedId:     fid,
		UnreadOnly: true,
	}))
	if err != nil {
		t.Fatalf("ListArticles after mark-all: %v", err)
	}
	if listResp.Msg.TotalCount != 0 {
		t.Errorf("expected 0 unread after mark-all, got %d", listResp.Msg.TotalCount)
	}
}

func TestMarkAllAsRead_FolderScope_MarksAllFeedsInFolder(t *testing.T) {
	db := testDB(t)
	feedStore := storage.NewFeedStore(db)
	folderStore := storage.NewFolderStore(db)
	articleStore := storage.NewArticleStore(db)
	svc := service.NewArticleService(articleStore)
	ctx := context.Background()

	folderID := mustCreateFolder(t, folderStore, "SidebarFolder")

	fid1 := ulid.Make().String()
	_, err := feedStore.Create(ctx, &storage.Feed{ID: fid1, Title: "F1", XMLURL: "http://f1.example/rss", FolderID: folderID, RefreshIntervalSeconds: 900})
	if err != nil {
		t.Fatalf("create feed1: %v", err)
	}
	fid2 := ulid.Make().String()
	_, err = feedStore.Create(ctx, &storage.Feed{ID: fid2, Title: "F2", XMLURL: "http://f2.example/rss", FolderID: folderID, RefreshIntervalSeconds: 900})
	if err != nil {
		t.Fatalf("create feed2: %v", err)
	}

	mustCreateArticle(t, articleStore, fid1, "a1")
	mustCreateArticle(t, articleStore, fid1, "a2")
	mustCreateArticle(t, articleStore, fid2, "b1")

	resp, err := svc.MarkAllAsRead(ctx, connect.NewRequest(&pb.MarkAllAsReadRequest{
		Scope: &pb.MarkAllAsReadRequest_FolderId{FolderId: folderID},
	}))
	if err != nil {
		t.Fatalf("MarkAllAsRead by folder: %v", err)
	}
	if resp.Msg.Count != 3 {
		t.Errorf("expected 3 marked, got %d", resp.Msg.Count)
	}
}
