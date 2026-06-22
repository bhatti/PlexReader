package storage

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// PreferencesStore defines the persistence interface for user preferences.
type PreferencesStore interface {
	Get(ctx context.Context) (*UserPreferences, error)
	Update(ctx context.Context, updates map[string]interface{}) (*UserPreferences, error)
}

type preferencesStore struct{ db *gorm.DB }

// NewPreferencesStore returns a PreferencesStore backed by GORM.
func NewPreferencesStore(db *gorm.DB) PreferencesStore { return &preferencesStore{db: db} }

func (s *preferencesStore) Get(ctx context.Context) (*UserPreferences, error) {
	var p UserPreferences
	if err := s.db.WithContext(ctx).First(&p, 1).Error; err != nil {
		return nil, fmt.Errorf("get preferences: %w", err)
	}
	return &p, nil
}

func (s *preferencesStore) Update(ctx context.Context, updates map[string]interface{}) (*UserPreferences, error) {
	if err := s.db.WithContext(ctx).Model(&UserPreferences{}).Where("id = 1").Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("update preferences: %w", err)
	}
	return s.Get(ctx)
}
