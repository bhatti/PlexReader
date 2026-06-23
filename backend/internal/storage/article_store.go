package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ListArticlesOpts configures article listing.
type ListArticlesOpts struct {
	FeedID        string
	FolderID      string
	UnreadOnly    bool
	ReadOnly      bool // only return read articles (for Recently Read)
	StarredOnly   bool
	SavedForLater bool
	TodayOnly     bool
	Query         string
	SortNewest    bool // true = newest first, false = oldest first
	PageSize      int
	PageToken     string
}

// MarkAllReadOpts configures bulk mark-as-read.
type MarkAllReadOpts struct {
	FeedID     string
	FolderID   string
	All        bool
	ArticleIDs []string
	Before     *time.Time
}

// ArticleStore defines the persistence interface for articles.
type ArticleStore interface {
	Create(ctx context.Context, a *Article) (*Article, error)
	BulkCreate(ctx context.Context, articles []*Article) (int, error)
	Get(ctx context.Context, id string) (*Article, error)
	List(ctx context.Context, opts ListArticlesOpts) ([]*Article, *PageResponse, error)
	MarkRead(ctx context.Context, id string, read bool) (*Article, error)
	MarkAllRead(ctx context.Context, opts MarkAllReadOpts) (int64, error)
	SetStarred(ctx context.Context, id string, starred bool) (*Article, error)
	SetSavedForLater(ctx context.Context, id string, saved bool) (*Article, error)
	DeleteByFeedID(ctx context.Context, feedID string) error
	DeleteExpired(ctx context.Context, retentionDays int) (int64, error)
}

type articleStore struct {
	db           *gorm.DB
	fts5Available bool
}

// NewArticleStore returns an ArticleStore backed by GORM.
// It probes for the articles_fts virtual table so full-text search degrades
// gracefully on SQLite builds that lack the FTS5 module.
func NewArticleStore(db *gorm.DB) ArticleStore {
	var count int64
	db.Raw("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='articles_fts'").Scan(&count)
	return &articleStore{db: db, fts5Available: count > 0}
}

func (s *articleStore) Create(ctx context.Context, a *Article) (*Article, error) {
	if err := s.db.WithContext(ctx).Create(a).Error; err != nil {
		return nil, fmt.Errorf("create article: %w", err)
	}
	return a, nil
}

// BulkCreate inserts articles, silently ignoring duplicates (same feed+guid).
// Returns the number of actually inserted rows.
// SQLite with CGO driver returns correct RowsAffected for ON CONFLICT DO NOTHING.
func (s *articleStore) BulkCreate(ctx context.Context, articles []*Article) (int, error) {
	if len(articles) == 0 {
		return 0, nil
	}
	// inserted is declared outside the transaction so the closure can accumulate
	// across batches without resetting on each call.
	var inserted int64
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		inserted = 0 // reset inside transaction in case of retry
		for i := 0; i < len(articles); i += 100 {
			end := i + 100
			if end > len(articles) {
				end = len(articles)
			}
			result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(articles[i:end])
			if result.Error != nil {
				return result.Error
			}
			inserted += result.RowsAffected
		}
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("bulk create articles: %w", err)
	}
	return int(inserted), nil
}

func (s *articleStore) Get(ctx context.Context, id string) (*Article, error) {
	var a Article
	if err := s.db.WithContext(ctx).First(&a, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("get article %s: %w", id, err)
	}
	return &a, nil
}

func (s *articleStore) List(ctx context.Context, opts ListArticlesOpts) ([]*Article, *PageResponse, error) {
	pageSize := NormalizePageSize(opts.PageSize)
	q := s.db.WithContext(ctx).Model(&Article{})

	if opts.Query != "" && s.fts5Available {
		// Full-text search via FTS5. Quote tokens to prevent query injection.
		q = q.Where("articles.rowid IN (SELECT rowid FROM articles_fts WHERE articles_fts MATCH ?)", sanitizeFTSQuery(opts.Query))
	}
	if opts.FeedID != "" {
		q = q.Where("feed_id = ?", opts.FeedID)
	}
	if opts.FolderID != "" {
		q = q.Joins("JOIN feeds ON feeds.id = articles.feed_id").
			Where("feeds.folder_id = ?", opts.FolderID)
	}
	if opts.UnreadOnly {
		q = q.Where("is_read = false")
	}
	if opts.ReadOnly {
		// "Recently Read" = individually opened articles only (read_at IS NOT NULL).
		// Articles bulk-marked via MarkAllRead do not have read_at set.
		q = q.Where("is_read = true AND read_at IS NOT NULL")
	}
	if opts.StarredOnly {
		q = q.Where("is_starred = true")
	}
	if opts.SavedForLater {
		q = q.Where("is_saved_for_later = true")
	}
	if opts.TodayOnly {
		q = q.Where("published_at >= datetime('now', '-1 day')")
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, nil, fmt.Errorf("count articles: %w", err)
	}

	// Always qualify with table name to avoid "ambiguous column" when a JOIN
	// (e.g. folder filter) is active.
	var order string
	if opts.ReadOnly {
		order = "articles.read_at desc, articles.id desc"
	} else if opts.SortNewest {
		order = "articles.published_at desc, articles.id desc"
	} else {
		order = "articles.published_at asc, articles.id asc"
	}

	if opts.PageToken != "" {
		c, err := DecodeCursor(opts.PageToken)
		if err != nil {
			return nil, nil, err
		}
		op := "<"
		if !opts.SortNewest {
			op = ">"
		}
		// COALESCE handles NULL published_at — treat as epoch so NULLs sort
		// consistently and cursors across NULL boundaries work correctly.
		q = q.Where(
			fmt.Sprintf("(COALESCE(articles.published_at, '1970-01-01T00:00:00Z'), articles.id) %s (?, ?)", op),
			c.SortValue, c.ID,
		)
	}

	var articles []*Article
	if err := q.Order(order).Limit(pageSize).Find(&articles).Error; err != nil {
		return nil, nil, fmt.Errorf("list articles: %w", err)
	}

	pr := &PageResponse{TotalCount: total}
	if len(articles) == pageSize {
		last := articles[len(articles)-1]
		var sv interface{}
		if last.PublishedAt != nil {
			sv = last.PublishedAt.Format(time.RFC3339)
		}
		pr.NextPageToken = EncodeCursor(sv, last.ID)
	}
	return articles, pr, nil
}

func (s *articleStore) MarkRead(ctx context.Context, id string, read bool) (*Article, error) {
	updates := map[string]interface{}{"is_read": read}
	if read {
		now := time.Now()
		updates["read_at"] = now
	} else {
		updates["read_at"] = nil
	}
	result := s.db.WithContext(ctx).Model(&Article{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return nil, fmt.Errorf("mark read %s: %w", id, result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("mark read %s: %w", id, gorm.ErrRecordNotFound)
	}
	return s.Get(ctx, id)
}

func (s *articleStore) MarkAllRead(ctx context.Context, opts MarkAllReadOpts) (int64, error) {
	q := s.db.WithContext(ctx).Model(&Article{}).Where("is_read = false")

	switch {
	case len(opts.ArticleIDs) > 0:
		q = q.Where("id IN ?", opts.ArticleIDs)
	case opts.FeedID != "":
		q = q.Where("feed_id = ?", opts.FeedID)
	case opts.FolderID != "":
		// SQLite UPDATE doesn't support JOIN — use a subquery instead.
		q = q.Where("feed_id IN (SELECT id FROM feeds WHERE folder_id = ?)", opts.FolderID)
	case opts.All:
		// no additional filter
	default:
		return 0, fmt.Errorf("mark all read: no scope specified")
	}

	if opts.Before != nil {
		q = q.Where("published_at < ?", opts.Before)
	}

	// Do NOT set read_at here — bulk mark-all-read should not pollute "Recently Read"
	// (which shows articles the user individually opened). Only MarkAsRead sets read_at.
	result := q.Updates(map[string]interface{}{"is_read": true})
	return result.RowsAffected, result.Error
}

func (s *articleStore) SetStarred(ctx context.Context, id string, starred bool) (*Article, error) {
	if err := s.db.WithContext(ctx).Model(&Article{}).Where("id = ?", id).
		Update("is_starred", starred).Error; err != nil {
		return nil, fmt.Errorf("set starred %s: %w", id, err)
	}
	return s.Get(ctx, id)
}

func (s *articleStore) SetSavedForLater(ctx context.Context, id string, saved bool) (*Article, error) {
	if err := s.db.WithContext(ctx).Model(&Article{}).Where("id = ?", id).
		Update("is_saved_for_later", saved).Error; err != nil {
		return nil, fmt.Errorf("set saved for later %s: %w", id, err)
	}
	return s.Get(ctx, id)
}

func (s *articleStore) DeleteByFeedID(ctx context.Context, feedID string) error {
	return s.db.WithContext(ctx).Delete(&Article{}, "feed_id = ?", feedID).Error
}

// sanitizeFTSQuery wraps each whitespace-delimited token in double quotes so
// that FTS5 treats them as phrase literals. This prevents users (or attackers)
// from injecting FTS5 syntax operators (AND, OR, NOT, NEAR, column:filter, *).
func sanitizeFTSQuery(q string) string {
	tokens := strings.Fields(q)
	if len(tokens) == 0 {
		return ""
	}
	for i, t := range tokens {
		// Escape internal double-quotes by doubling them (FTS5 quoting rule).
		tokens[i] = `"` + strings.ReplaceAll(t, `"`, `""`) + `"`
	}
	return strings.Join(tokens, " ")
}

// DeleteExpired removes read articles older than retentionDays.
// retentionDays=0 means keep forever.
func (s *articleStore) DeleteExpired(ctx context.Context, retentionDays int) (int64, error) {
	if retentionDays <= 0 {
		return 0, nil
	}
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	result := s.db.WithContext(ctx).
		Where("is_read = true AND is_starred = false AND is_saved_for_later = false AND created_at < ?", cutoff).
		Delete(&Article{})
	return result.RowsAffected, result.Error
}
