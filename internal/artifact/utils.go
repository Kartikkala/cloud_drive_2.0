package artifact

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/sirkartik/cloud_drive_2.0/internal/events"
	"github.com/sirkartik/cloud_drive_2.0/internal/storage"
)

func (svc *Service) downloadFile(
	ctx context.Context,
	Job *events.Job,
	WorkerID uint8,
) (*storage.Node, error) {
	stream, node, err := svc.StorageSvc.GetDataNoAuth(ctx, Job.NodeID)
	if err != nil {
		log.Println("err in GetDataNoAuth()")
		return nil, err
	}

	defer stream.Close()

	filename := fmt.Sprintf("videos/%v_%s", WorkerID, node.Name)
	file, err := os.Create(filename)
	defer file.Close()

	if err != nil {
		log.Println("err in os.Create()")
		os.Remove(filename)
		return nil, err
	}

	if _, err := io.Copy(file, stream); err != nil {
		log.Println("error in io.Copy()")
		os.Remove(filename)
		return nil, err
	}
	return node, nil
}
