package feed

import (
	"context"
	"crypto/sha256"
	"fmt"
	"html"
	"io"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/microcosm-cc/bluemonday"
	"github.com/mmcdole/gofeed"
)

// sanitizer uses bluemonday's UGC policy: allows common formatting tags but
// strips all JavaScript event handlers and dangerous attributes (href=javascript:,
// etc.) preventing stored XSS from malicious feed content.
var sanitizer = bluemonday.UGCPolicy()

// ParsedFeed is the normalized representation of any RSS or Atom feed.
type ParsedFeed struct {
	Title       string
	Description string
	Link        string
	IconURL     string
	Articles    []*ParsedArticle
}

// ParsedArticle is a single normalized feed entry.
type ParsedArticle struct {
	Title        string
	Link         string
	Content      string
	Summary      string
	Author       string
	PublishedAt  time.Time
	GUID         string
	ThumbnailURL string
}

// ParseFromReader parses an RSS/Atom feed from an io.Reader.
func ParseFromReader(r io.Reader) (*ParsedFeed, error) {
	fp := gofeed.NewParser()
	raw, err := fp.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("parse feed: %w", err)
	}
	return normalize(raw), nil
}

// ParseURL fetches and parses the feed at the given URL.
func ParseURL(ctx context.Context, url string) (*ParsedFeed, error) {
	fp := gofeed.NewParser()
	raw, err := fp.ParseURLWithContext(url, ctx)
	if err != nil {
		return nil, fmt.Errorf("parse feed url %s: %w", url, err)
	}
	return normalize(raw), nil
}

func normalize(raw *gofeed.Feed) *ParsedFeed {
	pf := &ParsedFeed{
		Title:       raw.Title,
		Description: raw.Description,
		Link:        raw.Link,
		IconURL:     extractFeedIcon(raw),
	}
	for _, item := range raw.Items {
		pf.Articles = append(pf.Articles, normalizeItem(item))
	}
	return pf
}

func normalizeItem(item *gofeed.Item) *ParsedArticle {
	a := &ParsedArticle{
		Title:        strings.TrimSpace(item.Title),
		Link:         item.Link,
		Content:      item.Content,
		Author:       extractAuthor(item),
		ThumbnailURL: extractThumbnail(item),
	}

	// Prefer full content over description for summary source.
	if a.Content == "" {
		a.Content = item.Description
	}
	// Sanitize HTML content to prevent stored XSS from malicious feeds.
	a.Content = sanitizer.Sanitize(a.Content)
	a.Summary = truncate(stripHTML(item.Description), 300)

	// Normalize published time.
	if item.PublishedParsed != nil {
		a.PublishedAt = *item.PublishedParsed
	} else if item.UpdatedParsed != nil {
		a.PublishedAt = *item.UpdatedParsed
	} else {
		a.PublishedAt = time.Now()
	}

	// Normalize GUID: prefer explicit GUID, fallback link, fallback hash.
	a.GUID = normalizeGUID(item)

	return a
}

func normalizeGUID(item *gofeed.Item) string {
	if item.GUID != "" {
		return item.GUID
	}
	if item.Link != "" {
		return item.Link
	}
	// Deterministic hash from title + published time.
	h := sha256.Sum256([]byte(item.Title + item.Published))
	return fmt.Sprintf("hash-%x", h[:8])
}

func extractAuthor(item *gofeed.Item) string {
	if item.Author != nil && item.Author.Name != "" {
		return item.Author.Name
	}
	if len(item.Authors) > 0 && item.Authors[0].Name != "" {
		return item.Authors[0].Name
	}
	return ""
}

func extractThumbnail(item *gofeed.Item) string {
	// Check media:thumbnail extension.
	if ext, ok := item.Extensions["media"]; ok {
		if thumbs, ok := ext["thumbnail"]; ok && len(thumbs) > 0 {
			if url, ok := thumbs[0].Attrs["url"]; ok {
				return url
			}
		}
		if content, ok := ext["content"]; ok && len(content) > 0 {
			if url, ok := content[0].Attrs["url"]; ok {
				return url
			}
		}
	}
	// Check enclosures for images.
	for _, enc := range item.Enclosures {
		if strings.HasPrefix(enc.Type, "image/") {
			return enc.URL
		}
	}
	return ""
}

func extractFeedIcon(raw *gofeed.Feed) string {
	if raw.Image != nil && raw.Image.URL != "" {
		return raw.Image.URL
	}
	return ""
}

// stripHTML removes HTML tags and decodes HTML entities from a string.
func stripHTML(s string) string {
	var b strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(html.UnescapeString(b.String()))
}

// truncate cuts s to at most n runes at a word boundary.
func truncate(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	runes := []rune(s)
	cut := runes[:n]
	// Find last space to avoid cutting mid-word.
	for i := len(cut) - 1; i > n/2; i-- {
		if cut[i] == ' ' {
			return string(cut[:i]) + "…"
		}
	}
	return string(cut) + "…"
}
