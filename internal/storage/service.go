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
	DB.AutoMigrate(&Node{})
	DB.AutoMigrate(&NodePermission{})
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
	node, err := svc.GetNode(ctx, NodeID)
	if err != nil || node.Type != NodeTypeDirectory || node.OwnerID != UserID  {
		return false
	}
	return true
}

func (svc *Service) checkNodeDeliverability(
	ctx context.Context,
	node *Node,
	UserID uint64,
) error {
	if node.Type != NodeTypeFile {
		return ErrNodeIsDirectory
	} else if node.OwnerID != UserID {
		return ErrUnauthorized
	}
	return nil
}

func (svc *Service) Put(ctx context.Context, 
	UserID uint64,
	ParentID uuid.UUID,
	Name string,
	Bytes uint64,
	data io.ReadCloser,
) error {
	defer data.Close()

	var parentID *uuid.UUID

	if ParentID != uuid.Nil {
		parentID = &ParentID
		if !svc.canWrite(ctx, ParentID, UserID) {
			return ErrUnauthorized
		}
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
			ParentID: parentID,
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

func (svc *Service) GetData(
	ctx context.Context,
	NodeID uuid.UUID,
	UserID uint64,
) (io.ReadCloser, *Node ,error) {
	node, err := svc.GetNode(ctx, NodeID)
	if err != nil {
		return nil, nil ,err
	}
	err = svc.checkNodeDeliverability(ctx, node, UserID)
	if err != nil {
		return nil, nil ,err
	}
	stream, err := svc.Client.Get(ctx, "cloud-drive", *node.Key)
	if err != nil {
		return nil, nil ,err
	}
	return stream, node, err
}

func (svc *Service) Delete(
	ctx context.Context,
	NodeID uuid.UUID,
	UserID uint64,
) error {
	node, err := svc.GetNode(ctx, NodeID)
	if err != nil {
		return err
	}
	err = svc.checkNodeDeliverability(ctx, node, UserID)
	if err != nil {
		return err
	}
	if errors.Is(err, ErrNodeIsDirectory) {
		// TODO : Recursive delete operation in directory
		return err
	} else if err != nil {
		return err
	}
	return svc.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Delete(&Node{} ,"id = ? AND owner_id = ?", NodeID, UserID)
		if res.Error != nil {
			return res.Error
		} 
		
		if res.RowsAffected == 0 {
			return ErrUnauthorized
		}

		if err := svc.Client.Delete(ctx, "cloud-drive", *node.Key); err != nil {
			return err
		}

		return nil
	})
}

func (svc Service) ListNodes(
	ctx context.Context,
	ParentNodeID uuid.UUID,
	UserID uint64,
) ([]Node,error) {
	
	var nodeList []Node
	db := svc.DB.WithContext(ctx).
    Where("owner_id = ?", UserID)

	if ParentNodeID == uuid.Nil {
		db = db.Where("parent_id IS NULL")
	} else {
		db = db.Where("parent_id = ?", ParentNodeID)
	}

	err := db.Find(&nodeList).Error

	if err != nil {
		return nil , err
	}
	return nodeList, nil
}

func (svc Service) CreateDirectoryNode(
	ctx context.Context,
	Name string,
	ParentNodeID uuid.UUID,
	OwnerID uint64,
) error {
	var parentId *uuid.UUID = nil
	if ParentNodeID != uuid.Nil {
		parentId = &ParentNodeID
	}
	var node Node = Node{
		Name: Name,
		ID: uuid.New(),
		ParentID: parentId,
		OwnerID: OwnerID,
		CreatedAt: time.Now(),
		Type: NodeTypeDirectory,
	}
	if err := svc.DB.WithContext(ctx).Create(&node).Error; err != nil {
		return err
	}
	return nil
}