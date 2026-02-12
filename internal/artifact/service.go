package artifact

import (
	"github.com/sirkartik/cloud_drive_2.0/internal/events"
	"github.com/sirkartik/cloud_drive_2.0/internal/storage"
	"gorm.io/gorm"
)

func NewService(DB *gorm.DB, StorageSvc *storage.Service, EventBroker *events.Broker[*events.Job], MaxNumberOfWorkers uint8) *Service {
	DB.AutoMigrate(&VideoArtifact{})
	DB.AutoMigrate(&VideoProcessingJob{})
	return &Service{
		DB:              DB,
		StorageSvc:      StorageSvc,
		ProcessingQueue: make(chan *events.Job, MaxNumberOfWorkers),
		MaxWorkers:      MaxNumberOfWorkers,
		EventBroker:     EventBroker,
	}
}
