package storage

import (
	"context"
	"errors"
	"io"
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func NewService(DB *gorm.DB, storageClient ObjectStorage) *Service {
	return &Service{
		DB : DB,
		Client: storageClient,
	}
}

func (svc *Service) GetNode(
	ctx context.Context,
	ID uuid.UUID,
)(*Node, error){
	var node Node
	err := svc.DB.WithContext(ctx).Where("id=?",ID).First(&node).Error
	if err != nil {
		return nil, err
	}
	return &node, nil
}

// TODO: Implement this

func (svc *Service) canWrite(
	ctx context.Context,
	NodeID uuid.UUID,
	UserID uint64,
) bool {
	return true
}

func (svc *Service) Put(ctx context.Context, 
	UserID uint64,
	ParentID uuid.UUID,
	Name string,
	Bytes uint64,
	data io.ReadCloser,
) error {
	defer data.Close()

	if !svc.canWrite(ctx, ParentID, UserID) {
		err := errors.New("Unauthorized write operation!")
		return err
	}
	mimeType, newReader , err := detectMimeType(data)
	if err != nil {
		return err
	}
	err = svc.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		key := uuid.NewString()
		node := Node{
			ID: uuid.New(),
			OwnerID: UserID,
			ParentID: &ParentID,
			Key: &key,
			CreatedAt: time.Now(),
			SizeBytes: &Bytes,
			MimeType: &mimeType,
			Type: NodeTypeFile,
			Name: Name,
		}

		result := tx.Create(&node)
		if result.Error != nil {
			return result.Error
		}
		err := svc.Client.Put(ctx, "cloud-drive", key, newReader, int64(Bytes))
		if err != nil {
			return err
		}
		
		return nil
	})
	return err
}