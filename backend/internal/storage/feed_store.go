package storage

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// ListFeedsOpts configures feed listing.
type ListFeedsOpts struct {
	FolderID  string
	PageSize  int
	PageToken string
}

// FeedStore defines the persistence interface for feeds.
type FeedStore interface {
	Create(ctx context.Context, feed *Feed) (*Feed, error)
	Get(ctx context.Context, id string) (*Feed, error)
	List(ctx context.Context, opts ListFeedsOpts) ([]*Feed, *PageResponse, error)
	Update(ctx context.Context, id string, updates map[string]interface{}) (*Feed, error)
	Delete(ctx context.Context, id string) error
	GetByXMLURL(ctx context.Context, xmlURL string) (*Feed, error)
	// UpdateLastFetched records a fetch attempt result.
	// backoff is how long to suppress future fetches (0 = no backoff, clear any existing).
	UpdateLastFetched(ctx context.Context, id string, t time.Time, lastErr string, backoff time.Duration) error
	// ListDueForRefresh returns feeds whose next fetch time has passed and whose
	// backoff window (if any) has expired.
	// globalMinSeconds is the floor interval applied to all feeds regardless of
	// their per-feed refresh_interval_seconds setting.
	ListDueForRefresh(ctx context.Context, globalMinSeconds int) ([]*Feed, error)
	GetUnreadCountsByFeed(ctx context.Context) (map[string]int64, error)
}

type feedStore struct{ db *gorm.DB }

// NewFeedStore returns a FeedStore backed by GORM.
func NewFeedStore(db *gorm.DB) FeedStore { return &feedStore{db: db} }

func (s *feedStore) Create(ctx context.Context, f *Feed) (*Feed, error) {
	if err := s.db.WithContext(ctx).Create(f).Error; err != nil {
		return nil, fmt.Errorf("create feed: %w", err)
	}
	return f, nil
}

func (s *feedStore) Get(ctx context.Context, id string) (*Feed, error) {
	var f Feed
	if err := s.db.WithContext(ctx).First(&f, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("get feed %s: %w", id, err)
	}
	return &f, nil
}

func (s *feedStore) List(ctx context.Context, opts ListFeedsOpts) ([]*Feed, *PageResponse, error) {
	pageSize := NormalizePageSize(opts.PageSize)
	q := s.db.WithContext(ctx).Model(&Feed{})
	if opts.FolderID != "" {
		q = q.Where("folder_id = ?", opts.FolderID)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, nil, fmt.Errorf("count feeds: %w", err)
	}

	if opts.PageToken != "" {
		c, err := DecodeCursor(opts.PageToken)
		if err != nil {
			return nil, nil, err
		}
		q = q.Where("(title, id) > (?, ?)", c.SortValue, c.ID)
	}

	var feeds []*Feed
	if err := q.Order("title asc, id asc").Limit(pageSize).Find(&feeds).Error; err != nil {
		return nil, nil, fmt.Errorf("list feeds: %w", err)
	}

	pr := &PageResponse{TotalCount: total}
	if len(feeds) == pageSize {
		last := feeds[len(feeds)-1]
		pr.NextPageToken = EncodeCursor(last.Title, last.ID)
	}
	return feeds, pr, nil
}

func (s *feedStore) Update(ctx context.Context, id string, updates map[string]interface{}) (*Feed, error) {
	if err := s.db.WithContext(ctx).Model(&Feed{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("update feed %s: %w", id, err)
	}
	return s.Get(ctx, id)
}

func (s *feedStore) Delete(ctx context.Context, id string) error {
	if err := s.db.WithContext(ctx).Delete(&Feed{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("delete feed %s: %w", id, err)
	}
	return nil
}

func (s *feedStore) GetByXMLURL(ctx context.Context, xmlURL string) (*Feed, error) {
	var f Feed
	err := s.db.WithContext(ctx).First(&f, "xml_url = ?", xmlURL).Error
	if err != nil {
		return nil, fmt.Errorf("get feed by url: %w", err)
	}
	return &f, nil
}

func (s *feedStore) UpdateLastFetched(ctx context.Context, id string, t time.Time, lastErr string, backoff time.Duration) error {
	updates := map[string]interface{}{
		"last_fetched_at": t,
		"last_error":      lastErr,
	}
	if lastErr == "" {
		updates["error_count"] = 0
		updates["backoff_until"] = nil
	} else {
		updates["error_count"] = gorm.Expr("error_count + 1")
		if backoff > 0 {
			bt := t.Add(backoff)
			updates["backoff_until"] = bt
		} else {
			updates["backoff_until"] = nil
		}
	}
	return s.db.WithContext(ctx).Model(&Feed{}).Where("id = ?", id).Updates(updates).Error
}

func (s *feedStore) ListDueForRefresh(ctx context.Context, globalMinSeconds int) ([]*Feed, error) {
	var feeds []*Feed
	// A feed is due when:
	//   1. Its refresh window has passed (or it has never been fetched), AND
	//   2. Its backoff window has expired (or there is no active backoff).
	err := s.db.WithContext(ctx).
		Where(`(last_fetched_at IS NULL OR datetime(last_fetched_at, '+' || MAX(refresh_interval_seconds, ?) || ' seconds') < datetime('now'))
		   AND (backoff_until IS NULL OR backoff_until < datetime('now'))`, globalMinSeconds).
		Find(&feeds).Error
	if err != nil {
		return nil, fmt.Errorf("list due for refresh: %w", err)
	}
	return feeds, nil
}

func (s *feedStore) GetUnreadCountsByFeed(ctx context.Context) (map[string]int64, error) {
	type row struct {
		FeedID string
		Count  int64
	}
	var rows []row
	err := s.db.WithContext(ctx).Model(&Article{}).
		Select("feed_id, count(*) as count").
		Where("is_read = false").
		Group("feed_id").
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("unread counts by feed: %w", err)
	}
	counts := make(map[string]int64, len(rows))
	for _, r := range rows {
		counts[r.FeedID] = r.Count
	}
	return counts, nil
}
