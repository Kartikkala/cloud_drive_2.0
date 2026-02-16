package storage

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"os"

	"github.com/google/uuid"
	"github.com/sirkartik/cloud_drive_2.0/internal/events"
	"gorm.io/gorm"
)

func NewService(DB *gorm.DB, storageClient ObjectStorage, EventBroker *events.Broker[*events.Job]) *Service {
	DB.AutoMigrate(&Node{})
	DB.AutoMigrate(&NodePermission{})
	return &Service{
		DB:          DB,
		Client:      storageClient,
		EventBroker: EventBroker,
	}
}

func (svc *Service) GetNode(
	ctx context.Context,
	ID uuid.UUID,
) (*Node, error) {
	var node Node
	err := svc.DB.WithContext(ctx).Where("id=?", ID).First(&node).Error
	if err != nil {
		return nil, err
	}
	return &node, nil
}

// TODO: Implement this

func (svc *Service) canWriteIntoDirectory(
	ctx context.Context,
	NodeID uuid.UUID,
	UserID uint64,
) error {
	node, err := svc.GetNode(ctx, NodeID)

	if err != nil {
		return err
	}

	if node.Type != NodeTypeDirectory {
		return ErrNodeIsFile
	} else if node.OwnerID != UserID {
		var permission NodePermission

		err = svc.DB.
			Where("user_id = ?", UserID).
			First(&permission).
			Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) || permission.Type != 2 || permission.Type != 5 || permission.Type != 7 {
				return ErrUnauthorized
			}
			return err
		}
	}
	return nil
}

func (svc *Service) checkNodeDeliverability(
	ctx context.Context,
	node *Node,
	UserID uint64,
) error {
	err := svc.DB.Model(&NodePermission{}).
		Where("user_id = ?", UserID).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) && UserID != node.OwnerID {
			return ErrUnauthorized
		}
		return err
	} else if node.Type != NodeTypeFile {
		return ErrNodeIsDirectory
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
		if err := svc.canWriteIntoDirectory(ctx, ParentID, UserID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrParentNodeNotFound
			}
			return err
		}
	}

	mimeType, newReader, err := detectMimeType(data)
	if err != nil {
		return err
	}
	err = svc.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		key := uuid.NewString()
		nodeID := uuid.New()
		node := Node{
			ID:        nodeID,
			OwnerID:   UserID,
			ParentID:  parentID,
			Key:       &key,
			CreatedAt: time.Now(),
			SizeBytes: &Bytes,
			MimeType:  &mimeType,
			Type:      NodeTypeFile,
			Name:      Name,
		}

		result := tx.Create(&node)
		if result.Error != nil {
			return result.Error
		}
		err := svc.Client.Put(ctx, "cloud-drive", key, newReader, int64(Bytes))
		if err != nil {
			return err
		}

		if strings.HasPrefix(mimeType, "video/") {
			job := &events.Job{
				NodeID: nodeID,
			}
			svc.EventBroker.Publish("video", job)
		}

		return nil
	})
	return err
}

func (svc *Service) GetData(
	ctx context.Context,
	NodeID uuid.UUID,
	UserID uint64,
) (io.ReadCloser, *Node, error) {
	node, err := svc.GetNode(ctx, NodeID)
	if err != nil {
		return nil, nil, err
	}
	err = svc.checkNodeDeliverability(ctx, node, UserID)
	if err != nil {
		return nil, nil, err
	}
	stream, err := svc.Client.Get(ctx, "cloud-drive", *node.Key)
	if err != nil {
		return nil, nil, err
	}
	return stream, node, err
}

func (svc *Service) Delete(
	ctx context.Context,
	NodeID uuid.UUID,
	UserID uint64,
) error {

	// Fetch all matching nodes
	var nodes []Node
	err := svc.DB.WithContext(ctx).
		Raw(`
		WITH RECURSIVE subtree AS (
		SELECT * FROM nodes WHERE id = ?
		AND owner_id = ?
		
		UNION ALL
		
		SELECT n.* FROM nodes n JOIN
		subtree s ON n.parent_id = s.id
		)
	
		SELECT * FROM subtree;
	`, NodeID, UserID).
		Scan(&nodes).Error

	if err != nil {
		return err
	}

	if len(nodes) == 0 {
		return ErrNodeNotFound
	}

	// Deletion from object storage

	for _, item := range nodes {
		if item.Key == nil {
			continue
		}
		svc.Client.Delete(ctx, "cloud-drive", *item.Key)
	}

	return svc.DB.WithContext(ctx).
		Exec(`
	WITH RECURSIVE subtree AS (
		SELECT id
		FROM nodes
		WHERE id = ? AND owner_id = ?

		UNION ALL

		SELECT n.id
		FROM nodes n
		JOIN subtree s ON n.parent_id = s.id
	)
	DELETE FROM nodes
	WHERE id IN (SELECT id FROM subtree);
	`, NodeID, UserID).
		Error
}

// No AUTH version for Get for Internal Services

func (svc *Service) GetDataNoAuth(
	ctx context.Context,
	NodeID uuid.UUID,
) (io.ReadCloser, *Node, error) {
	node, err := svc.GetNode(ctx, NodeID)
	if err != nil {
		return nil, nil, err
	}
	stream, err := svc.Client.Get(ctx, "cloud-drive", *node.Key)
	if err != nil {
		return nil, nil, err
	}
	return stream, node, err
}

func (svc Service) ListNodes(
	ctx context.Context,
	ParentNodeID uuid.UUID,
	UserID uint64,
) ([]NodeWithPermission, error) {

	var nodeList []NodeWithPermission

	db := svc.DB.WithContext(ctx).
		Table("nodes").
		Select("nodes.*, node_permissions.type AS permission_type").
		Joins(`
			LEFT JOIN node_permissions 
			ON node_permissions.node_id = nodes.id 
			AND node_permissions.user_id = ?
		`, UserID).
		Where("nodes.owner_id = ? OR node_permissions.user_id = ?", UserID, UserID)

	if ParentNodeID == uuid.Nil {
		db = db.Where("nodes.parent_id IS NULL")
	} else {
		db = db.Where("nodes.parent_id = ?", ParentNodeID)
	}

	err := db.Scan(&nodeList).Error
	if err != nil {
		return nil, err
	}

	return nodeList, nil
}

// TODO 1. Delete Playlist function, list
// artifacts function with permission
// matrix

func (svc Service) PutHLS(
	ctx context.Context,
	HLSDirPath,
	ParentKey string,
) error {
	if _, err := os.Stat(HLSDirPath); err != nil {
		return err
	}
	// TODO : Revert in case of error
	err := filepath.WalkDir(HLSDirPath, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}

			defer file.Close()

			info, err := d.Info()

			if err != nil {
				return err
			}

			relpath, err := filepath.Rel(HLSDirPath, path)

			if err != nil {
				return err
			}

			relpath = filepath.ToSlash(relpath)
			key := strings.Join([]string{ParentKey, relpath}, "/")

			err = svc.Client.Put(
				ctx,
				"cloud-drive-hls",
				key,
				file,
				info.Size(),
			)

			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
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
		Name:      Name,
		ID:        uuid.New(),
		ParentID:  parentId,
		OwnerID:   OwnerID,
		CreatedAt: time.Now(),
		Type:      NodeTypeDirectory,
	}

	return svc.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if ParentNodeID != uuid.Nil {
			tx = tx.Where("id = ? AND owner_id = ? AND type = 'directory'", ParentNodeID, OwnerID)
			var parentNode Node
			if err := tx.First(&parentNode).Error; err != nil {
				return err
			}
		}
		if err := tx.Create(&node).Error; err != nil {
			return err
		}
		return nil
	})
}

func (svc Service) Copy(
	ctx context.Context,
	TargetNodeID uuid.UUID,
	DestinationID uuid.UUID,
	OwnerID uint64,
) error {
	if TargetNodeID == uuid.Nil {
		return errors.New("target node id can't be nil")
	}

	db := svc.DB.WithContext(ctx)

	var targetNode Node
	if err := db.Where("id = ? AND owner_id = ?", TargetNodeID, OwnerID).
		First(&targetNode).Error; err != nil {
		return err
	}

	if targetNode.Type == NodeTypeDirectory {
		return errors.New("directory copy not supported")
	}

	if DestinationID != uuid.Nil {
		isDes, err := svc.isDescendant(ctx, TargetNodeID, DestinationID, OwnerID)
		if err != nil {
			return err
		}
		if isDes {
			return errors.New("cannot copy node into its own subtree")
		}
	}

	if DestinationID != uuid.Nil {
		var destinationNode Node
		if err := db.Where("id = ? AND owner_id = ?", DestinationID, OwnerID).
			First(&destinationNode).Error; err != nil {
			return err
		}
		if destinationNode.Type == NodeTypeFile {
			return errors.New("cannot copy into a file")
		}
	}

	newNodeKey := uuid.NewString()

	if err := svc.Client.Copy(ctx, "cloud-drive", *targetNode.Key, newNodeKey); err != nil {
		return err
	}

	var destinationID *uuid.UUID
	if DestinationID != uuid.Nil {
		destinationID = &DestinationID
	}

	newNode := Node{
		ID:        uuid.New(),
		ParentID:  destinationID,
		OwnerID:   OwnerID,
		Name:      targetNode.Name,
		Type:      targetNode.Type,
		Key:       &newNodeKey,
		SizeBytes: targetNode.SizeBytes,
		MimeType:  targetNode.MimeType,
		CreatedAt: time.Now(),
	}

	if err := svc.DB.WithContext(ctx).Create(&newNode).Error; err != nil {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = svc.Client.Delete(cleanupCtx, "cloud-drive", newNodeKey)
		return err
	}

	return nil
}

func (svc Service) isDescendant(
	ctx context.Context,
	TargetNodeID uuid.UUID,
	DestinationNodeID uuid.UUID,
	OwnerID uint64,
) (bool, error) {
	var found int = 0

	err := svc.DB.WithContext(ctx).
		Raw(`
		WITH RECURSIVE subtree AS (
		SELECT id FROM public.nodes
		WHERE parent_id = ?
		AND owner_id = ?

		UNION ALL

		SELECT n.id
		FROM subtree s JOIN public.nodes n ON s.id = n.parent_id
		)
		SELECT 1 FROM subtree WHERE id = ? LIMIT 1;
	`, TargetNodeID, OwnerID, DestinationNodeID).Scan(&found).Error

	if err != nil {
		return false, err
	}

	return found == 1, nil
}

func (svc Service) Move(
	ctx context.Context,
	TargetNodeID uuid.UUID,
	DestinationParentID uuid.UUID,
	OwnerID uint64,
) error {
	if TargetNodeID == uuid.Nil {
		return errors.New("target node id can't be nil")
	}
	var destId *uuid.UUID = nil

	if DestinationParentID != uuid.Nil {
		isDes, err := svc.isDescendant(ctx, TargetNodeID, DestinationParentID, OwnerID)

		if err != nil {
			return err
		}

		if isDes {
			return errors.New("cannot move node into its own subtree")
		}
		destId = &DestinationParentID
	}

	query := svc.DB.WithContext(ctx).
		Model(&Node{}).
		Where("owner_id = ?", OwnerID)

	query = query.Where("id = ?", TargetNodeID)
	result := query.Update("parent_id", destId)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("node not found or source parent mismatch")
	}
	return nil
}
