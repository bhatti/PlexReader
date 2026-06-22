package service

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/oklog/ulid/v2"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/plexreader/plexreader/backend/gen/plexreader/v1"
	"github.com/plexreader/plexreader/backend/gen/plexreader/v1/plexreaderv1connect"
	"github.com/plexreader/plexreader/backend/internal/storage"
)

// FolderService implements plexreaderv1connect.FolderServiceHandler.
type FolderService struct {
	store storage.FolderStore
}

func NewFolderService(store storage.FolderStore) plexreaderv1connect.FolderServiceHandler {
	return &FolderService{store: store}
}

func (s *FolderService) CreateFolder(ctx context.Context, req *connect.Request[pb.CreateFolderRequest]) (*connect.Response[pb.Folder], error) {
	f := req.Msg.Folder
	if f == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("folder is required"))
	}
	folder := &storage.Folder{
		ID:       ulid.Make().String(),
		Name:     f.Name,
		ParentID: f.ParentId,
	}
	created, err := s.store.Create(ctx, folder)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(folderToProto(created, 0)), nil
}

func (s *FolderService) GetFolder(ctx context.Context, req *connect.Request[pb.GetFolderRequest]) (*connect.Response[pb.Folder], error) {
	f, err := s.store.Get(ctx, req.Msg.Id)
	if err != nil {
		return nil, notFoundOrInternal(err)
	}
	counts, _ := s.store.GetUnreadCounts(ctx)
	return connect.NewResponse(folderToProto(f, int32(counts[f.ID]))), nil
}

func (s *FolderService) ListFolders(ctx context.Context, req *connect.Request[pb.ListFoldersRequest]) (*connect.Response[pb.ListFoldersResponse], error) {
	folders, err := s.store.List(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	counts, _ := s.store.GetUnreadCounts(ctx)
	protos := make([]*pb.Folder, len(folders))
	for i, f := range folders {
		protos[i] = folderToProto(f, int32(counts[f.ID]))
	}
	return connect.NewResponse(&pb.ListFoldersResponse{Folders: protos}), nil
}

func (s *FolderService) UpdateFolder(ctx context.Context, req *connect.Request[pb.UpdateFolderRequest]) (*connect.Response[pb.Folder], error) {
	f := req.Msg.Folder
	if f == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("folder is required"))
	}
	updates := map[string]interface{}{
		"name":      f.Name,
		"parent_id": f.ParentId,
		"position":  f.Position,
	}
	updated, err := s.store.Update(ctx, f.Id, updates)
	if err != nil {
		return nil, notFoundOrInternal(err)
	}
	counts, _ := s.store.GetUnreadCounts(ctx)
	return connect.NewResponse(folderToProto(updated, int32(counts[updated.ID]))), nil
}

func (s *FolderService) DeleteFolder(ctx context.Context, req *connect.Request[pb.DeleteFolderRequest]) (*connect.Response[emptypb.Empty], error) {
	if err := s.store.Delete(ctx, req.Msg.Id); err != nil {
		return nil, notFoundOrInternal(err)
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (s *FolderService) ReorderFolders(ctx context.Context, req *connect.Request[pb.ReorderFoldersRequest]) (*connect.Response[pb.ReorderFoldersResponse], error) {
	folders, err := s.store.Reorder(ctx, req.Msg.FolderIds)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	protos := make([]*pb.Folder, len(folders))
	for i, f := range folders {
		protos[i] = folderToProto(f, 0)
	}
	return connect.NewResponse(&pb.ReorderFoldersResponse{Folders: protos}), nil
}
