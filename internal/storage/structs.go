package storage

import (
	"context"
	"io"
	"net/url"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/sirkartik/cloud_drive_2.0/internal/config"
	"github.com/sirkartik/cloud_drive_2.0/internal/shared"
	"gorm.io/gorm"
)

type Service struct {
	DB     *gorm.DB
	Client shared.ObjectStorage
	Cfg    config.Config
}

type Handler struct {
	svc StorageService
}

type StorageService interface {
	GetNode(ctx context.Context, ID uuid.UUID) (*Node, error)
	DetectMimeType(ctx context.Context, data io.ReadCloser) (string, io.ReadCloser, error)
	Put(ctx context.Context, UserID uint64, ParentID uuid.UUID, Name string, Bytes uint64, data io.ReadCloser, mimeType string) (*Node, error)
	GetData(ctx context.Context, NodeID uuid.UUID, UserID uint64) (io.ReadCloser, *Node, error)
	GeneratePresignedGetURL(ctx context.Context, key string) (*url.URL, error)
	Delete(ctx context.Context, NodeID uuid.UUID, UserID uint64) error
	GetDataNoAuth(ctx context.Context, NodeID uuid.UUID) (io.ReadCloser, *Node, error)
	ListNodes(ctx context.Context, ParentNodeID uuid.UUID, UserID uint64) ([]NodeWithPermission, error)
	PutHLS(ctx context.Context, HLSDirPath, ParentKey string) error
	CreateDirectoryNode(ctx context.Context, Name string, ParentNodeID uuid.UUID, OwnerID uint64) error
	Copy(ctx context.Context, TargetNodeID uuid.UUID, DestinationID uuid.UUID, OwnerID uint64) error
	Move(ctx context.Context, TargetNodeID uuid.UUID, DestinationParentID uuid.UUID, OwnerID uint64) error
	GeneratePostUploadPolicy(ctx context.Context) (*UploadPolicy, error)
}

type HookLayer struct {
	storageSvc    StorageService
	putHooksAfter []PutHook
}

type PutHook func(
	ctx context.Context,
	userID uint64,
	parentID uuid.UUID,
	fileName string,
	mimeType string,
	nodeId uuid.UUID,
	key string,
	sizeBytes uint64,
) error

type NodeWithPermission struct {
	Node
	PermissionType *PermissionType
}

type MinioStorage struct {
	client *minio.Client
}

type UploadPolicy struct {
	URL       string
	Fields    map[string]string
	KeyPrefix string
}
