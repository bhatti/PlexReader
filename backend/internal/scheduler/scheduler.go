package scheduler

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog"

	"github.com/plexreader/plexreader/backend/internal/feed"
	"github.com/plexreader/plexreader/backend/internal/storage"
)

const defaultConcurrency = 4 // default max concurrent feed fetches

// maxErrorsBeforeDelete is the number of consecutive permanent errors (DNS
// NXDOMAIN, HTTP 404/410, etc.) after which the feed is automatically removed.
const maxErrorsBeforeDelete = 5

// Scheduler periodically refreshes feeds and prunes expired articles.
type Scheduler struct {
	feedStore    storage.FeedStore
	articleStore storage.ArticleStore
	prefStore    storage.PreferencesStore
	fetcher      *feed.Fetcher
	logger       zerolog.Logger
	interval     time.Duration
	concurrency  int
	cancel       context.CancelFunc
	wg           sync.WaitGroup
}

// NewForTest creates a minimal Scheduler for unit tests (no fetcher, no interval).
func NewForTest(
	feedStore storage.FeedStore,
	articleStore storage.ArticleStore,
	prefStore storage.PreferencesStore,
) *Scheduler {
	return &Scheduler{
		feedStore:    feedStore,
		articleStore: articleStore,
		prefStore:    prefStore,
		logger:       zerolog.Nop(),
	}
}

// RunRetention executes one retention-cleanup pass. Exported for tests.
func (s *Scheduler) RunRetention(ctx context.Context) { s.runRetention(ctx) }

// New creates a Scheduler.
func New(
	feedStore storage.FeedStore,
	articleStore storage.ArticleStore,
	prefStore storage.PreferencesStore,
	fetcher *feed.Fetcher,
	interval time.Duration,
	concurrency int,
	logger zerolog.Logger,
) *Scheduler {
	if concurrency <= 0 {
		concurrency = defaultConcurrency
	}
	return &Scheduler{
		feedStore:    feedStore,
		articleStore: articleStore,
		prefStore:    prefStore,
		fetcher:      fetcher,
		logger:       logger,
		interval:     interval,
		concurrency:  concurrency,
	}
}

// Concurrency returns the configured max concurrent fetch goroutines.
func (s *Scheduler) Concurrency() int { return s.concurrency }

// Start begins the background refresh loop and the hourly retention cleanup.
func (s *Scheduler) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	// Feed refresh loop.
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.runCycle(ctx)
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.runCycle(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()

	// Retention cleanup — runs immediately on start, then every hour.
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.runRetention(ctx)
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.runRetention(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()
}

// Stop cancels in-flight fetches and waits for the loop to drain.
func (s *Scheduler) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
	s.wg.Wait()
}

// RefreshFeed manually triggers a refresh for a single feed by ID.
func (s *Scheduler) RefreshFeed(ctx context.Context, feedID string) error {
	f, err := s.feedStore.Get(ctx, feedID)
	if err != nil {
		return err
	}
	return s.refreshOne(ctx, f)
}

func (s *Scheduler) runRetention(ctx context.Context) {
	prefs, err := s.prefStore.Get(ctx)
	if err != nil {
		s.logger.Error().Err(err).Msg("retention: load preferences")
		return
	}
	days := prefs.RetentionDays
	if days <= 0 {
		days = 90
	}
	deleted, err := s.articleStore.DeleteExpired(ctx, days)
	if err != nil {
		s.logger.Error().Err(err).Msg("retention cleanup")
		return
	}
	if deleted > 0 {
		s.logger.Info().Int64("deleted", deleted).Int("days", days).Msg("retention: deleted old articles")
	}
}

func (s *Scheduler) runCycle(ctx context.Context) {
	// Load preferences for the global refresh floor.
	prefs, err := s.prefStore.Get(ctx)
	if err != nil {
		s.logger.Error().Err(err).Msg("load preferences")
		return
	}
	globalMin := prefs.GlobalRefreshIntervalSeconds
	if globalMin <= 0 {
		globalMin = 3600
	}

	feeds, err := s.feedStore.ListDueForRefresh(ctx, globalMin)
	if err != nil {
		s.logger.Error().Err(err).Msg("list due for refresh")
		return
	}
	if len(feeds) == 0 {
		return
	}
	s.logger.Info().Int("count", len(feeds)).Msg("refreshing feeds")

	sem := make(chan struct{}, s.concurrency)
	var wg sync.WaitGroup
	for _, f := range feeds {
		f := f
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			if err := s.refreshOne(ctx, f); err != nil {
				permanent, backoff := classifyError(err)
				s.logger.Warn().Err(err).
					Str("feed", f.ID).
					Str("url", f.XMLURL).
					Bool("permanent", permanent).
					Dur("backoff", backoff).
					Msg("refresh failed")
				if permanent {
					// Re-read current error count (refreshOne already incremented it).
					updated, dbErr := s.feedStore.Get(ctx, f.ID)
					if dbErr == nil {
						if updated.ErrorCount >= maxErrorsBeforeDelete {
							s.logger.Warn().
								Str("feed", f.Title).
								Str("url", f.XMLURL).
								Int("errors", updated.ErrorCount).
								Msg("removing permanently failed feed")
							_ = s.articleStore.DeleteByFeedID(ctx, f.ID)
							_ = s.feedStore.Delete(ctx, f.ID)
						} else {
							s.logger.Warn().
								Str("feed", f.Title).
								Str("url", f.XMLURL).
								Int("errors", updated.ErrorCount).
								Int("remove_after", maxErrorsBeforeDelete-updated.ErrorCount).
								Msg("permanent error — feed will be removed after more failures")
						}
					}
				}
			}
		}()
	}
	wg.Wait()
}

// classifyError returns whether the error is permanent (auto-delete after N
// failures) and how long to back off before the next retry.
//
// Error classes:
//   - 404 / 410 / DNS NXDOMAIN / unparseable non-feed: permanent — remove after N hits.
//   - 403 Forbidden: transient — back off 1 hour, never auto-remove.
//   - 429 Too Many Requests: transient — back off 5 minutes, never auto-remove.
//   - Other (5xx, network timeout, etc.): transient — no explicit backoff,
//     rely on the normal refresh interval.
func classifyError(err error) (permanent bool, backoff time.Duration) {
	var httpErr *feed.HTTPError
	if errors.As(err, &httpErr) {
		switch httpErr.StatusCode {
		case 404, 410:
			return true, 0
		case 403:
			return false, time.Hour
		case 429:
			return false, 5 * time.Minute
		}
		return false, 0
	}
	msg := err.Error()
	if strings.Contains(msg, "no such host") {
		return true, 0
	}
	if strings.Contains(msg, "Failed to detect feed type") {
		return true, 0
	}
	return false, 0
}

func (s *Scheduler) refreshOne(ctx context.Context, f *storage.Feed) error {
	result, err := s.fetcher.Fetch(ctx, f.XMLURL)
	fetchedAt := time.Now()

	if err != nil {
		_, backoff := classifyError(err)
		_ = s.feedStore.UpdateLastFetched(ctx, f.ID, fetchedAt, err.Error(), backoff)
		return err
	}
	if result.NotModified {
		_ = s.feedStore.UpdateLastFetched(ctx, f.ID, fetchedAt, "", 0)
		return nil
	}

	articles := make([]*storage.Article, 0, len(result.Feed.Articles))
	for _, pa := range result.Feed.Articles {
		articles = append(articles, &storage.Article{
			ID:           ulid.Make().String(),
			FeedID:       f.ID,
			GUIDFeedID:   f.ID,
			Title:        pa.Title,
			Link:         pa.Link,
			Content:      pa.Content,
			Summary:      pa.Summary,
			Author:       pa.Author,
			PublishedAt:  &pa.PublishedAt,
			GUID:         pa.GUID,
			ThumbnailURL: pa.ThumbnailURL,
		})
	}

	created, err := s.articleStore.BulkCreate(ctx, articles)
	if err != nil {
		_ = s.feedStore.UpdateLastFetched(ctx, f.ID, fetchedAt, err.Error(), 0)
		return err
	}

	_ = s.feedStore.UpdateLastFetched(ctx, f.ID, fetchedAt, "", 0)
	if created > 0 {
		s.logger.Info().
			Str("feed", f.Title).
			Int("new_articles", created).
			Msg("feed refreshed")
	}
	return nil
}
