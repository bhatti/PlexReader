package storage

import "time"

// Folder organizes feeds into named groups.
type Folder struct {
	ID        string    `gorm:"primaryKey"`
	Name      string    `gorm:"not null"`
	ParentID  string    `gorm:"default:''"`
	Position  int       `gorm:"default:0"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// Feed represents a subscribed RSS/Atom source.
type Feed struct {
	ID                     string     `gorm:"primaryKey"`
	Title                  string     `gorm:"not null"`
	XMLURL                 string     `gorm:"uniqueIndex;not null;column:xml_url"`
	HTMLURL                string     `gorm:"column:html_url"`
	Description            string
	IconURL                string     `gorm:"column:icon_url"`
	FolderID               string     `gorm:"index"`
	RefreshIntervalSeconds int        `gorm:"default:900"`
	LastFetchedAt          *time.Time `gorm:"index"`
	// BackoffUntil is set on transient errors (403, 429) to suppress retries
	// until the backoff window expires.  NULL means no active backoff.
	BackoffUntil *time.Time `gorm:"index"`
	IsFavorite   bool      `gorm:"default:false"`
	LastError    string
	ErrorCount   int       `gorm:"default:0"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}

// Article is a single item fetched from a Feed.
type Article struct {
	ID              string     `gorm:"primaryKey"`
	FeedID          string     `gorm:"index;not null"`
	Title           string
	Link            string
	Content         string
	Summary         string
	Author          string
	PublishedAt     *time.Time `gorm:"index"`
	GUID            string     `gorm:"uniqueIndex:idx_article_feed_guid"`
	GUIDFeedID      string     `gorm:"uniqueIndex:idx_article_feed_guid;column:guid_feed_id"`
	ThumbnailURL    string     `gorm:"column:thumbnail_url"`
	IsRead          bool       `gorm:"index;default:false"`
	IsStarred       bool       `gorm:"index;default:false"`
	IsSavedForLater bool       `gorm:"index;default:false;column:is_saved_for_later"`
	ReadAt          *time.Time `gorm:"index"`
	CreatedAt       time.Time `gorm:"autoCreateTime"`
}

// UserPreferences holds global reader settings. Single row (id=1).
type UserPreferences struct {
	ID                           int    `gorm:"primaryKey"`
	StartPage                    string `gorm:"default:'today'"`
	DefaultView                  string `gorm:"default:'magazine'"`
	DefaultSort                  string `gorm:"default:'newest_first'"`
	HideReadArticles             bool   `gorm:"default:true"`
	GlobalRefreshIntervalSeconds int    `gorm:"default:3600"`
	RetentionDays                int    `gorm:"default:90"`
	Theme                        string `gorm:"default:'dark'"`
}
