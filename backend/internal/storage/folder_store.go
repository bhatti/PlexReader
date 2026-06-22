package storage

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// FolderStore defines the persistence interface for folders.
type FolderStore interface {
	Create(ctx context.Context, folder *Folder) (*Folder, error)
	Get(ctx context.Context, id string) (*Folder, error)
	List(ctx context.Context) ([]*Folder, error)
	Update(ctx context.Context, id string, updates map[string]interface{}) (*Folder, error)
	Delete(ctx context.Context, id string) error
	Reorder(ctx context.Context, ids []string) ([]*Folder, error)
	GetUnreadCounts(ctx context.Context) (map[string]int64, error)
}

type folderStore struct{ db *gorm.DB }

// NewFolderStore returns a FolderStore backed by GORM.
func NewFolderStore(db *gorm.DB) FolderStore { return &folderStore{db: db} }

func (s *folderStore) Create(ctx context.Context, f *Folder) (*Folder, error) {
	if err := s.db.WithContext(ctx).Create(f).Error; err != nil {
		return nil, fmt.Errorf("create folder: %w", err)
	}
	return f, nil
}

func (s *folderStore) Get(ctx context.Context, id string) (*Folder, error) {
	var f Folder
	if err := s.db.WithContext(ctx).First(&f, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("get folder %s: %w", id, err)
	}
	return &f, nil
}

func (s *folderStore) List(ctx context.Context) ([]*Folder, error) {
	var folders []*Folder
	if err := s.db.WithContext(ctx).Order("position asc, name asc").Find(&folders).Error; err != nil {
		return nil, fmt.Errorf("list folders: %w", err)
	}
	return folders, nil
}

func (s *folderStore) Update(ctx context.Context, id string, updates map[string]interface{}) (*Folder, error) {
	if err := s.db.WithContext(ctx).Model(&Folder{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("update folder %s: %w", id, err)
	}
	return s.Get(ctx, id)
}

func (s *folderStore) Delete(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Unset folder_id on feeds before deleting the folder (atomic).
		if err := tx.Model(&Feed{}).Where("folder_id = ?", id).
			Update("folder_id", "").Error; err != nil {
			return fmt.Errorf("unset feeds folder: %w", err)
		}
		if err := tx.Delete(&Folder{}, "id = ?", id).Error; err != nil {
			return fmt.Errorf("delete folder %s: %w", id, err)
		}
		return nil
	})
}

func (s *folderStore) Reorder(ctx context.Context, ids []string) ([]*Folder, error) {
	var folders []*Folder
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			result := tx.Model(&Folder{}).Where("id = ?", id).Update("position", i)
			if result.Error != nil {
				return fmt.Errorf("reorder folder %s: %w", id, result.Error)
			}
			if result.RowsAffected == 0 {
				return fmt.Errorf("reorder folder %s: %w", id, gorm.ErrRecordNotFound)
			}
		}
		return tx.Order("position asc").Find(&folders).Error
	})
	return folders, err
}

func (s *folderStore) GetUnreadCounts(ctx context.Context) (map[string]int64, error) {
	type result struct {
		FolderID string
		Count    int64
	}
	var rows []result
	err := s.db.WithContext(ctx).
		Model(&Article{}).
		Select("feeds.folder_id, count(*) as count").
		Joins("JOIN feeds ON feeds.id = articles.feed_id").
		Where("articles.is_read = false AND feeds.folder_id != ''").
		Group("feeds.folder_id").
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("get unread counts: %w", err)
	}
	counts := make(map[string]int64, len(rows))
	for _, r := range rows {
		counts[r.FolderID] = r.Count
	}
	return counts, nil
}
