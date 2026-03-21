package hooks

import (
	"context"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

func NewArtifactsSvcHooks(nc *nats.Conn) *ArtifactsSvcHooks {
	return &ArtifactsSvcHooks{
		nc: nc,
	}
}

func (svc *ArtifactsSvcHooks) OnVideo(
	ctx context.Context,
	userID uint64,
	parentID uuid.UUID,
	fileName string,
	mimeType string,
	sizeBytes uint64,
) error {
	if strings.HasPrefix(mimeType, "video/") {
		log.Printf("New video file %s, invoking artifacts svc...", fileName)
		return svc.nc.Publish(
			"video",
			[]byte("data"),
		)
	}
	return nil
}
