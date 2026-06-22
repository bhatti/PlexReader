package storage

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
)

const (
	defaultPageSize = 50
	maxPageSize     = 200
)

// PageRequest holds pagination parameters from the caller.
type PageRequest struct {
	PageSize  int
	PageToken string
}

// PageResponse holds pagination state to return to the caller.
type PageResponse struct {
	NextPageToken string
	TotalCount    int64
}

// cursor encodes the last-seen row for cursor-based pagination.
type cursor struct {
	SortValue interface{} `json:"sv"`
	ID        string      `json:"id"`
}

// EncodeCursor encodes a sort value + ID into a page token.
func EncodeCursor(sortValue interface{}, id string) string {
	c := cursor{SortValue: sortValue, ID: id}
	b, _ := json.Marshal(c)
	return base64.StdEncoding.EncodeToString(b)
}

// DecodeCursor decodes a page token back into a cursor.
func DecodeCursor(token string) (*cursor, error) {
	b, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return nil, fmt.Errorf("invalid page token: %w", err)
	}
	var c cursor
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, fmt.Errorf("invalid page token: %w", err)
	}
	return &c, nil
}

// NormalizePageSize clamps the requested page size to valid bounds.
func NormalizePageSize(requested int) int {
	if requested <= 0 {
		return defaultPageSize
	}
	if requested > maxPageSize {
		return maxPageSize
	}
	return requested
}

// CountQuery executes a count on the given query and returns total.
func CountQuery(db *gorm.DB) (int64, error) {
	var count int64
	return count, db.Count(&count).Error
}
