package service

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	pb "github.com/plexreader/plexreader/backend/gen/plexreader/v1"
	"github.com/plexreader/plexreader/backend/gen/plexreader/v1/plexreaderv1connect"
	"github.com/plexreader/plexreader/backend/internal/storage"
)

// ArticleService implements plexreaderv1connect.ArticleServiceHandler.
type ArticleService struct {
	store storage.ArticleStore
}

func NewArticleService(store storage.ArticleStore) plexreaderv1connect.ArticleServiceHandler {
	return &ArticleService{store: store}
}

func (s *ArticleService) GetArticle(ctx context.Context, req *connect.Request[pb.GetArticleRequest]) (*connect.Response[pb.Article], error) {
	a, err := s.store.Get(ctx, req.Msg.Id)
	if err != nil {
		return nil, notFoundOrInternal(err)
	}
	return connect.NewResponse(articleToProto(a)), nil
}

func (s *ArticleService) ListArticles(ctx context.Context, req *connect.Request[pb.ListArticlesRequest]) (*connect.Response[pb.ListArticlesResponse], error) {
	opts := storage.ListArticlesOpts{
		FeedID:        req.Msg.FeedId,
		FolderID:      req.Msg.FolderId,
		UnreadOnly:    req.Msg.UnreadOnly,
		ReadOnly:      req.Msg.ReadOnly,
		StarredOnly:   req.Msg.StarredOnly,
		SavedForLater: req.Msg.SavedForLaterOnly,
		TodayOnly:     req.Msg.TodayOnly,
		Query:         req.Msg.Query,
		SortNewest:    req.Msg.SortOrder != pb.SortOrder_SORT_ORDER_OLDEST_FIRST,
		PageSize:      int(req.Msg.PageSize),
		PageToken:     req.Msg.PageToken,
	}
	articles, pr, err := s.store.List(ctx, opts)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	protos := make([]*pb.Article, len(articles))
	for i, a := range articles {
		protos[i] = articleToProto(a)
	}
	return connect.NewResponse(&pb.ListArticlesResponse{
		Articles:      protos,
		NextPageToken: pr.NextPageToken,
		TotalCount:    int32(pr.TotalCount),
	}), nil
}

func (s *ArticleService) MarkAsRead(ctx context.Context, req *connect.Request[pb.MarkAsReadRequest]) (*connect.Response[pb.Article], error) {
	a, err := s.store.MarkRead(ctx, req.Msg.Id, true)
	if err != nil {
		return nil, notFoundOrInternal(err)
	}
	return connect.NewResponse(articleToProto(a)), nil
}

func (s *ArticleService) MarkAsUnread(ctx context.Context, req *connect.Request[pb.MarkAsUnreadRequest]) (*connect.Response[pb.Article], error) {
	a, err := s.store.MarkRead(ctx, req.Msg.Id, false)
	if err != nil {
		return nil, notFoundOrInternal(err)
	}
	return connect.NewResponse(articleToProto(a)), nil
}

func (s *ArticleService) MarkAllAsRead(ctx context.Context, req *connect.Request[pb.MarkAllAsReadRequest]) (*connect.Response[pb.MarkAllAsReadResponse], error) {
	opts := storage.MarkAllReadOpts{
		ArticleIDs: req.Msg.ArticleIds,
		Before:     timePtr(req.Msg.BeforeTime),
	}
	switch scope := req.Msg.Scope.(type) {
	case *pb.MarkAllAsReadRequest_FeedId:
		opts.FeedID = scope.FeedId
	case *pb.MarkAllAsReadRequest_FolderId:
		opts.FolderID = scope.FolderId
	case *pb.MarkAllAsReadRequest_All:
		opts.All = scope.All
	default:
		if len(opts.ArticleIDs) == 0 {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("scope or article_ids required"))
		}
	}
	count, err := s.store.MarkAllRead(ctx, opts)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&pb.MarkAllAsReadResponse{Count: int32(count)}), nil
}

func (s *ArticleService) StarArticle(ctx context.Context, req *connect.Request[pb.StarArticleRequest]) (*connect.Response[pb.Article], error) {
	a, err := s.store.SetStarred(ctx, req.Msg.Id, req.Msg.Starred)
	if err != nil {
		return nil, notFoundOrInternal(err)
	}
	return connect.NewResponse(articleToProto(a)), nil
}

func (s *ArticleService) SaveForLater(ctx context.Context, req *connect.Request[pb.SaveForLaterRequest]) (*connect.Response[pb.Article], error) {
	a, err := s.store.SetSavedForLater(ctx, req.Msg.Id, req.Msg.Saved)
	if err != nil {
		return nil, notFoundOrInternal(err)
	}
	return connect.NewResponse(articleToProto(a)), nil
}
