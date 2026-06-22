package service

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/plexreader/plexreader/backend/gen/plexreader/v1"
	"github.com/plexreader/plexreader/backend/internal/storage"
)

func folderToProto(f *storage.Folder, unreadCount int32) *pb.Folder {
	return &pb.Folder{
		Id:          f.ID,
		Name:        f.Name,
		ParentId:    f.ParentID,
		Position:    int32(f.Position),
		UnreadCount: unreadCount,
		CreateTime:  timestamppb.New(f.CreatedAt),
		UpdateTime:  timestamppb.New(f.UpdatedAt),
	}
}

func feedToProto(f *storage.Feed, unreadCount int32) *pb.Feed {
	pf := &pb.Feed{
		Id:                     f.ID,
		Title:                  f.Title,
		XmlUrl:                 f.XMLURL,
		HtmlUrl:                f.HTMLURL,
		FolderId:               f.FolderID,
		Description:            f.Description,
		IconUrl:                f.IconURL,
		RefreshIntervalSeconds: int32(f.RefreshIntervalSeconds),
		UnreadCount:            unreadCount,
		LastError:              f.LastError,
		ErrorCount:             int32(f.ErrorCount),
		CreateTime:             timestamppb.New(f.CreatedAt),
		UpdateTime:             timestamppb.New(f.UpdatedAt),
	}
	if f.LastFetchedAt != nil {
		pf.LastFetchedTime = timestamppb.New(*f.LastFetchedAt)
	}
	return pf
}

func articleToProto(a *storage.Article) *pb.Article {
	pa := &pb.Article{
		Id:              a.ID,
		FeedId:          a.FeedID,
		Title:           a.Title,
		Link:            a.Link,
		Content:         a.Content,
		Summary:         a.Summary,
		Author:          a.Author,
		Guid:            a.GUID,
		ThumbnailUrl:    a.ThumbnailURL,
		IsRead:          a.IsRead,
		IsStarred:       a.IsStarred,
		IsSavedForLater: a.IsSavedForLater,
		CreateTime:      timestamppb.New(a.CreatedAt),
	}
	if a.PublishedAt != nil {
		pa.PublishedTime = timestamppb.New(*a.PublishedAt)
	}
	if a.ReadAt != nil {
		pa.ReadAt = timestamppb.New(*a.ReadAt)
	}
	return pa
}

func prefsToProto(p *storage.UserPreferences) *pb.UserPreferences {
	return &pb.UserPreferences{
		StartPage:                    stringToStartPage(p.StartPage),
		DefaultView:                  stringToViewMode(p.DefaultView),
		DefaultSort:                  stringToSortOrder(p.DefaultSort),
		HideReadArticles:             p.HideReadArticles,
		GlobalRefreshIntervalSeconds: int32(p.GlobalRefreshIntervalSeconds),
		RetentionDays:                int32(p.RetentionDays),
		Theme:                        p.Theme,
	}
}

func timePtr(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	t := ts.AsTime()
	return &t
}

func stringToStartPage(s string) pb.StartPage {
	switch s {
	case "today":
		return pb.StartPage_START_PAGE_TODAY
	case "first_folder":
		return pb.StartPage_START_PAGE_FIRST_FOLDER
	case "all":
		return pb.StartPage_START_PAGE_ALL
	}
	return pb.StartPage_START_PAGE_TODAY
}

func startPageToString(sp pb.StartPage) string {
	switch sp {
	case pb.StartPage_START_PAGE_FIRST_FOLDER:
		return "first_folder"
	case pb.StartPage_START_PAGE_ALL:
		return "all"
	default:
		return "today"
	}
}

func stringToViewMode(s string) pb.ViewMode {
	switch s {
	case "title_only":
		return pb.ViewMode_VIEW_MODE_TITLE_ONLY
	case "magazine":
		return pb.ViewMode_VIEW_MODE_MAGAZINE
	case "cards":
		return pb.ViewMode_VIEW_MODE_CARDS
	case "article":
		return pb.ViewMode_VIEW_MODE_ARTICLE
	}
	return pb.ViewMode_VIEW_MODE_MAGAZINE
}

func viewModeToString(vm pb.ViewMode) string {
	switch vm {
	case pb.ViewMode_VIEW_MODE_TITLE_ONLY:
		return "title_only"
	case pb.ViewMode_VIEW_MODE_CARDS:
		return "cards"
	case pb.ViewMode_VIEW_MODE_ARTICLE:
		return "article"
	default:
		return "magazine"
	}
}

func stringToSortOrder(s string) pb.SortOrder {
	if s == "oldest_first" {
		return pb.SortOrder_SORT_ORDER_OLDEST_FIRST
	}
	return pb.SortOrder_SORT_ORDER_NEWEST_FIRST
}

func sortOrderToString(so pb.SortOrder) string {
	if so == pb.SortOrder_SORT_ORDER_OLDEST_FIRST {
		return "oldest_first"
	}
	return "newest_first"
}
