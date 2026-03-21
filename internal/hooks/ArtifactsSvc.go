package hooks

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/sirkartik/cloud_drive_2.0/internal/storage"
)

func NewArtifactsSvcHooks(storageSvc *storage.Service, nc *nats.Conn) *ArtifactsSvcHooks {
	return &ArtifactsSvcHooks{
		storageSvc: storageSvc,
		nc:         nc,
	}
}

func (svc *ArtifactsSvcHooks) OnVideo(
	ctx context.Context,
	userID uint64,
	parentID uuid.UUID,
	fileName string,
	mimeType string,
	nodeID uuid.UUID,
	key string,
	sizeBytes uint64,
) error {
	if strings.HasPrefix(mimeType, "video/") {
		log.Printf("New video file %s, invoking artifacts svc...", fileName)
		url, err := svc.storageSvc.GeneratePresignedGetURL(ctx, key)
		if err != nil {
			return err
		}

		videoJob := &VideoJob{
			URL:    url.String(),
			NodeID: nodeID.String(),
		}

		payload, err := json.Marshal(videoJob)
		if err != nil {
			return err
		}
		return svc.nc.Publish(
			"video.new",
			payload,
		)
	}
	return nil
}
