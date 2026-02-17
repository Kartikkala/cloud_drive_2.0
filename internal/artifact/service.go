package artifact

import (
	"log"

	"github.com/sirkartik/cloud_drive_2.0/internal/events"
	"github.com/sirkartik/cloud_drive_2.0/internal/storage"
	"gorm.io/gorm"
)

func NewService(
	DB *gorm.DB, StorageSvc *storage.Service,
	NewJobEventBroker *events.Broker[*events.Job],
	ProgressUpdateEventBroker *events.Broker[*events.JobProgress],
	MaxNumberOfWorkers uint8,
) *Service {
	encoder := encoderForVendor(DetectGPUVendor())

	log.Println("Selected encoder : ", encoder)
	DB.AutoMigrate(&VideoArtifact{})
	DB.AutoMigrate(&VideoProcessingJob{})
	return &Service{
		DB:                   DB,
		StorageSvc:           StorageSvc,
		ProcessingQueue:      make(chan *events.Job, MaxNumberOfWorkers),
		MaxWorkers:           MaxNumberOfWorkers,
		NewJobEventBroker:    NewJobEventBroker,
		ProgressUpdateBroker: ProgressUpdateEventBroker,
		VideoEncoder:         encoder,
	}
}
