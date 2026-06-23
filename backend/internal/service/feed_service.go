package service

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/oklog/ulid/v2"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/plexreader/plexreader/backend/gen/plexreader/v1"
	"github.com/plexreader/plexreader/backend/gen/plexreader/v1/plexreaderv1connect"
	"github.com/plexreader/plexreader/backend/internal/feed"
	"github.com/plexreader/plexreader/backend/internal/scheduler"
	"github.com/plexreader/plexreader/backend/internal/storage"
)

// FeedService implements plexreaderv1connect.FeedServiceHandler.
type FeedService struct {
	feedStore   storage.FeedStore
	folderStore storage.FolderStore
	sched       *scheduler.Scheduler
}

func NewFeedService(feedStore storage.FeedStore, folderStore storage.FolderStore, sched *scheduler.Scheduler) plexreaderv1connect.FeedServiceHandler {
	return &FeedService{feedStore: feedStore, folderStore: folderStore, sched: sched}
}

func (s *FeedService) CreateFeed(ctx context.Context, req *connect.Request[pb.CreateFeedRequest]) (*connect.Response[pb.Feed], error) {
	f := req.Msg.Feed
	if f == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("feed is required"))
	}
	if err := feed.ValidateFeedURL(f.XmlUrl); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	stored := &storage.Feed{
		ID:                     ulid.Make().String(),
		Title:                  f.Title,
		XMLURL:                 f.XmlUrl,
		HTMLURL:                f.HtmlUrl,
		FolderID:               f.FolderId,
		Description:            f.Description,
		IconURL:                f.IconUrl,
		IsFavorite:             f.IsFavorite,
		RefreshIntervalSeconds: int(f.RefreshIntervalSeconds),
	}
	if stored.RefreshIntervalSeconds == 0 {
		stored.RefreshIntervalSeconds = 900
	}
	created, err := s.feedStore.Create(ctx, stored)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	// Trigger an immediate background refresh so articles appear right away.
	go func() {
		if err := s.sched.RefreshFeed(context.Background(), created.ID); err != nil {
			// Best-effort; the scheduler will retry on the next cycle.
			_ = err
		}
	}()
	return connect.NewResponse(feedToProto(created, 0)), nil
}

func (s *FeedService) GetFeed(ctx context.Context, req *connect.Request[pb.GetFeedRequest]) (*connect.Response[pb.Feed], error) {
	f, err := s.feedStore.Get(ctx, req.Msg.Id)
	if err != nil {
		return nil, notFoundOrInternal(err)
	}
	counts, _ := s.feedStore.GetUnreadCountsByFeed(ctx)
	return connect.NewResponse(feedToProto(f, int32(counts[f.ID]))), nil
}

func (s *FeedService) ListFeeds(ctx context.Context, req *connect.Request[pb.ListFeedsRequest]) (*connect.Response[pb.ListFeedsResponse], error) {
	opts := storage.ListFeedsOpts{
		FolderID:  req.Msg.FolderId,
		PageSize:  int(req.Msg.PageSize),
		PageToken: req.Msg.PageToken,
	}
	feeds, pr, err := s.feedStore.List(ctx, opts)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	counts, _ := s.feedStore.GetUnreadCountsByFeed(ctx)
	protos := make([]*pb.Feed, len(feeds))
	for i, f := range feeds {
		protos[i] = feedToProto(f, int32(counts[f.ID]))
	}
	return connect.NewResponse(&pb.ListFeedsResponse{
		Feeds:         protos,
		NextPageToken: pr.NextPageToken,
	}), nil
}

func (s *FeedService) UpdateFeed(ctx context.Context, req *connect.Request[pb.UpdateFeedRequest]) (*connect.Response[pb.Feed], error) {
	f := req.Msg.Feed
	if f == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("feed is required"))
	}
	updates := map[string]interface{}{
		"title":                    f.Title,
		"html_url":                 f.HtmlUrl,
		"folder_id":                f.FolderId,
		"description":              f.Description,
		"icon_url":                 f.IconUrl,
		"refresh_interval_seconds": f.RefreshIntervalSeconds,
		"is_favorite":              f.IsFavorite,
	}
	updated, err := s.feedStore.Update(ctx, f.Id, updates)
	if err != nil {
		return nil, notFoundOrInternal(err)
	}
	counts, _ := s.feedStore.GetUnreadCountsByFeed(ctx)
	return connect.NewResponse(feedToProto(updated, int32(counts[updated.ID]))), nil
}

func (s *FeedService) DeleteFeed(ctx context.Context, req *connect.Request[pb.DeleteFeedRequest]) (*connect.Response[emptypb.Empty], error) {
	if err := s.feedStore.Delete(ctx, req.Msg.Id); err != nil {
		return nil, notFoundOrInternal(err)
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (s *FeedService) ImportOPML(ctx context.Context, req *connect.Request[pb.ImportOPMLRequest]) (*connect.Response[pb.ImportOPMLResponse], error) {
	doc, err := feed.ParseOPML(req.Msg.OpmlContent)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	var foldersCreated, feedsCreated, feedsSkipped int32
	var errs []string

	// Load existing folders once to avoid N queries during import.
	existingFolders, err := s.folderStore.List(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	folderMap := make(map[string]string, len(existingFolders))
	for _, f := range existingFolders {
		folderMap[f.Name] = f.ID
	}

	// Create missing folders, collecting folder name → ID mapping.
	for _, of := range doc.Folders {
		if _, exists := folderMap[of.Name]; !exists {
			f := &storage.Folder{ID: ulid.Make().String(), Name: of.Name}
			created, err := s.folderStore.Create(ctx, f)
			if err != nil {
				errs = append(errs, fmt.Sprintf("create folder %q: %v", of.Name, err))
				continue
			}
			folderMap[of.Name] = created.ID
			foldersCreated++
		}
		// Create feeds in this folder.
		for _, of2 := range of.Feeds {
			n, err := s.importFeed(ctx, of2, folderMap[of.Name])
			feedsCreated += int32(n)
			if err != nil {
				feedsSkipped++
				errs = append(errs, fmt.Sprintf("feed %q: %v", of2.XMLURL, err))
			}
		}
	}
	// Uncategorized feeds.
	for _, of := range doc.Feeds {
		n, err := s.importFeed(ctx, of, "")
		feedsCreated += int32(n)
		if err != nil {
			feedsSkipped++
			errs = append(errs, fmt.Sprintf("feed %q: %v", of.XMLURL, err))
		}
	}
	return connect.NewResponse(&pb.ImportOPMLResponse{
		FeedsCreated:   feedsCreated,
		FoldersCreated: foldersCreated,
		FeedsSkipped:   feedsSkipped,
		Errors:         errs,
	}), nil
}

func (s *FeedService) ExportOPML(ctx context.Context, req *connect.Request[pb.ExportOPMLRequest]) (*connect.Response[pb.ExportOPMLResponse], error) {
	// Paginate through all feeds rather than relying on a single large-page
	// request — MaxPageSize caps at 200, so 1000 was silently truncated before.
	var allFeeds []*storage.Feed
	pageToken := ""
	for {
		page, pr, err := s.feedStore.List(ctx, storage.ListFeedsOpts{PageSize: 200, PageToken: pageToken})
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		allFeeds = append(allFeeds, page...)
		if pr.NextPageToken == "" {
			break
		}
		pageToken = pr.NextPageToken
	}
	allFolders, err := s.folderStore.List(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	doc := &feed.OPMLDocument{Title: "PlexReader Export"}
	folderMap := make(map[string]*feed.OPMLFolder)
	for _, f := range allFolders {
		of := &feed.OPMLFolder{Name: f.Name}
		doc.Folders = append(doc.Folders, of)
		folderMap[f.ID] = of
	}
	for _, f := range allFeeds {
		of := &feed.OPMLFeed{Title: f.Title, XMLURL: f.XMLURL, HTMLURL: f.HTMLURL}
		if folder, ok := folderMap[f.FolderID]; ok {
			folder.Feeds = append(folder.Feeds, of)
		} else {
			doc.Feeds = append(doc.Feeds, of)
		}
	}

	data, err := feed.GenerateOPML(doc)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&pb.ExportOPMLResponse{OpmlContent: data}), nil
}

func (s *FeedService) RefreshFeed(ctx context.Context, req *connect.Request[pb.RefreshFeedRequest]) (*connect.Response[pb.Feed], error) {
	if err := s.sched.RefreshFeed(ctx, req.Msg.Id); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return s.GetFeed(ctx, connect.NewRequest(&pb.GetFeedRequest{Id: req.Msg.Id}))
}

func (s *FeedService) importFeed(ctx context.Context, of *feed.OPMLFeed, folderID string) (int, error) {
	if of.XMLURL == "" {
		return 0, fmt.Errorf("missing xml_url")
	}
	// Skip if already subscribed.
	existing, err := s.feedStore.GetByXMLURL(ctx, of.XMLURL)
	if err == nil {
		_ = existing
		return 0, nil // already exists — not an error, just skip
	}
	if !isNotFound(err) {
		// Real storage error — propagate so the caller can report it.
		return 0, fmt.Errorf("check existing feed: %w", err)
	}
	f := &storage.Feed{
		ID:                     ulid.Make().String(),
		Title:                  of.Title,
		XMLURL:                 of.XMLURL,
		HTMLURL:                of.HTMLURL,
		FolderID:               folderID,
		RefreshIntervalSeconds: 900,
	}
	created, err := s.feedStore.Create(ctx, f)
	if err != nil {
		return 0, err
	}
	// Trigger an immediate background refresh so articles appear right away.
	if s.sched != nil {
		go func() {
			if err := s.sched.RefreshFeed(context.Background(), created.ID); err != nil {
				_ = err // scheduler will retry on next cycle
			}
		}()
	}
	return 1, nil
}
