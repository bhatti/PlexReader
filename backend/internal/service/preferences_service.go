package service

import (
	"context"

	"connectrpc.com/connect"

	pb "github.com/plexreader/plexreader/backend/gen/plexreader/v1"
	"github.com/plexreader/plexreader/backend/gen/plexreader/v1/plexreaderv1connect"
	"github.com/plexreader/plexreader/backend/internal/storage"
)

// PreferencesService implements plexreaderv1connect.PreferencesServiceHandler.
type PreferencesService struct {
	store storage.PreferencesStore
}

func NewPreferencesService(store storage.PreferencesStore) plexreaderv1connect.PreferencesServiceHandler {
	return &PreferencesService{store: store}
}

func (s *PreferencesService) GetPreferences(ctx context.Context, req *connect.Request[pb.GetPreferencesRequest]) (*connect.Response[pb.UserPreferences], error) {
	p, err := s.store.Get(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(prefsToProto(p)), nil
}

func (s *PreferencesService) UpdatePreferences(ctx context.Context, req *connect.Request[pb.UpdatePreferencesRequest]) (*connect.Response[pb.UserPreferences], error) {
	p := req.Msg.Preferences
	if p == nil {
		p = &pb.UserPreferences{}
	}

	updates := map[string]interface{}{
		"start_page":                      startPageToString(p.StartPage),
		"default_view":                    viewModeToString(p.DefaultView),
		"default_sort":                    sortOrderToString(p.DefaultSort),
		"hide_read_articles":              p.HideReadArticles,
		"global_refresh_interval_seconds": p.GlobalRefreshIntervalSeconds,
		"retention_days":                  p.RetentionDays,
	}
	if p.Theme != "" {
		updates["theme"] = p.Theme
	}

	updated, err := s.store.Update(ctx, updates)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(prefsToProto(updated)), nil
}
