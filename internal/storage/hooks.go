package storage

import (
	"context"
	"io"
	"log"
	"net/url"

	"github.com/google/uuid"
)

func NewHookLayer(storageSvc StorageService) *HookLayer {
	return &HookLayer{
		storageSvc:    storageSvc,
		putHooksAfter: []PutHook{},
	}
}

func (h *HookLayer) RegisterAfterPutHook(hook PutHook) {
	h.putHooksAfter = append(h.putHooksAfter, hook)
}

func (h *HookLayer) Put(
	ctx context.Context,
	UserID uint64,
	ParentID uuid.UUID,
	Name string,
	Bytes uint64,
	data io.ReadCloser,
	mimeType string,
) (*Node, error) {
	node, err := h.storageSvc.Put(ctx, UserID, ParentID, Name, Bytes, data, mimeType)
	if err != nil {
		return nil, err
	}

	for _, hook := range h.putHooksAfter {
		var parentId uuid.UUID
		if node.ParentID != nil {
			parentId = *node.ParentID
		}

		err := hook(ctx, UserID, parentId, Name, mimeType, node.ID, *node.Key, Bytes)
		if err != nil {
			log.Println("AfterPut hook error : ", err)
		}
	}

	return node, nil
}

func (h *HookLayer) GetNode(ctx context.Context, ID uuid.UUID) (*Node, error) {
	return h.storageSvc.GetNode(ctx, ID)
}

func (h *HookLayer) DetectMimeType(ctx context.Context, data io.ReadCloser) (string, io.ReadCloser, error) {
	return h.storageSvc.DetectMimeType(ctx, data)
}

func (h *HookLayer) GetData(ctx context.Context, NodeID uuid.UUID, UserID uint64) (io.ReadCloser, *Node, error) {
	return h.storageSvc.GetData(ctx, NodeID, UserID)
}

func (h *HookLayer) GeneratePresignedGetURL(ctx context.Context, key string) (*url.URL, error) {
	return h.storageSvc.GeneratePresignedGetURL(ctx, key)
}

func (h *HookLayer) Delete(ctx context.Context, NodeID uuid.UUID, UserID uint64) error {
	return h.storageSvc.Delete(ctx, NodeID, UserID)
}

func (h *HookLayer) GetDataNoAuth(ctx context.Context, NodeID uuid.UUID) (io.ReadCloser, *Node, error) {
	return h.storageSvc.GetDataNoAuth(ctx, NodeID)
}

func (h *HookLayer) ListNodes(ctx context.Context, ParentNodeID uuid.UUID, UserID uint64) ([]NodeWithPermission, error) {
	return h.storageSvc.ListNodes(ctx, ParentNodeID, UserID)
}

func (h *HookLayer) PutHLS(ctx context.Context, HLSDirPath, ParentKey string) error {
	return h.storageSvc.PutHLS(ctx, HLSDirPath, ParentKey)
}

func (h *HookLayer) CreateDirectoryNode(ctx context.Context, Name string, ParentNodeID uuid.UUID, OwnerID uint64) error {
	return h.storageSvc.CreateDirectoryNode(ctx, Name, ParentNodeID, OwnerID)
}

func (h *HookLayer) Copy(ctx context.Context, TargetNodeID uuid.UUID, DestinationID uuid.UUID, OwnerID uint64) error {
	return h.storageSvc.Copy(ctx, TargetNodeID, DestinationID, OwnerID)
}

func (h *HookLayer) Move(ctx context.Context, TargetNodeID uuid.UUID, DestinationParentID uuid.UUID, OwnerID uint64) error {
	return h.storageSvc.Move(ctx, TargetNodeID, DestinationParentID, OwnerID)
}

func (h *HookLayer) GeneratePostUploadPolicy(ctx context.Context) (*UploadPolicy, error) {
	return h.storageSvc.GeneratePostUploadPolicy(ctx)
}


