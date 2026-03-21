package hooks

import (
	"github.com/nats-io/nats.go"
	"github.com/sirkartik/cloud_drive_2.0/internal/storage"
)

type ArtifactsSvcHooks struct {
	storageSvc *storage.Service
	nc         *nats.Conn
}

type VideoJob struct {
	NodeID string `json:"node_id"`
	URL    string `json:"url"`
}
