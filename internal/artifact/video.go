package artifact

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/google/uuid"
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

func (svc Service) StartWorkers(
	ctx context.Context,
) {
	go svc.jobProducer(ctx)
	for i := uint8(0); i < svc.MaxWorkers; i++ {
		go svc.videoWorker(ctx, i)
	}
}

func (svc *Service) jobProducer(
	ctx context.Context,
) {
	videoCh := svc.EventBroker.Subscribe("video")
	// TODO: Publish A job.completed event on job completion
	jobCompletedCh := svc.EventBroker.Subscribe("job.completed")

	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-videoCh:
			if !ok {
				return
			}
			log.Println("fired video event!")
			svc.fillQueue(ctx, job)
		case <-jobCompletedCh:
			log.Println("fired job completed!")
			svc.fillQueue(ctx, nil)
		}
	}
}

func (svc *Service) fillQueue(
	ctx context.Context,
	job *events.Job,
) {
	if len(svc.ProcessingQueue) >= cap(svc.ProcessingQueue) {
		return
	}

	var videoProcessingJob VideoProcessingJob
	if job != nil {
		videoProcessingJob = VideoProcessingJob{
			NodeID:     job.NodeID,
			Status:     StatusPending,
			CreatedAt:  time.Now(),
			AccessedAt: time.Now(),
		}
		if err := svc.DB.WithContext(ctx).
			Create(&videoProcessingJob).Error; err != nil {
			log.Println("error in job registration", err)
			return
		}
	} else {
		err := svc.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := tx.Raw(`
        		SELECT * FROM video_processing_jobs
        		WHERE status = ?
        		ORDER BY accessed_at ASC, created_at ASC
        		FOR UPDATE SKIP LOCKED
        		LIMIT 1`, StatusPending).
				Scan(&videoProcessingJob).Error; err != nil {
				return err
			}

			if videoProcessingJob.NodeID == uuid.Nil {
				return gorm.ErrRecordNotFound
			}

			return tx.Model(&VideoProcessingJob{}).
				Where("node_id = ?", videoProcessingJob.NodeID).
				Updates(map[string]any{
					"status":      StatusProcessing,
					"accessed_at": time.Now(),
				}).Error
		})

		if err != nil {
			log.Println("error in job fetch: ", err)
			return
		}
	}

	select {
	case svc.ProcessingQueue <- &events.Job{NodeID: videoProcessingJob.NodeID}:
		return
	case <-ctx.Done():
		return
	default:
		return
	}
}

func (svc *Service) ffprobe(
	ctx context.Context,
	FilePath string,
) (*VideoMetadata, error) {
	cmd := exec.CommandContext(ctx,
		"ffprobe",
		"-v", "error",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		FilePath)

	out, err := cmd.Output()

	if err != nil {
		return nil, err
	}

	var ffprobeOutput FFprobeOutput
	var videoMetadata VideoMetadata
	videoMetadata.HasAudio = false
	err = json.Unmarshal(out, &ffprobeOutput)

	if err != nil {
		return nil, err
	}

	for _, stream := range ffprobeOutput.Streams {
		if stream.CodecType == "audio" {
			videoMetadata.HasAudio = true
		} else if stream.CodecType == "video" {
			videoMetadata.Bitrate = ffprobeOutput.Format.Bitrate
			videoMetadata.Codec = stream.CodecName
			videoMetadata.Duration = ffprobeOutput.Format.Duration
			videoMetadata.Height = stream.Height
			videoMetadata.Width = stream.Width
		}
	}

	return &videoMetadata, nil
}

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

func (svc *Service) setJobStatusFailed(
	ctx context.Context,
	job *events.Job,
) error {
	if err := svc.DB.WithContext(ctx).
		Model(&VideoProcessingJob{}).
		Where("node_id = ?", job.NodeID).
		Update("status", StatusFailed).Error; err != nil {
		return err
	}
	return nil
}

func (svc *Service) videoWorker(
	ctx context.Context,
	WorkerID uint8,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case job := <-svc.ProcessingQueue:
			node, err := svc.downloadFile(ctx, job, WorkerID)

			if err != nil {
				log.Println("error in video worker (downloading file):", err)
				err = svc.setJobStatusFailed(ctx, job)
				svc.EventBroker.Publish("job.completed", job)
				continue
			}

			filename := fmt.Sprintf("videos/%v_%s", WorkerID, node.Name)
			vm, err := svc.ffprobe(ctx, filename)
			os.Remove(filename)
			if err != nil {
				log.Println("error in video worker: (ffprobe)", err)
				err = svc.setJobStatusFailed(ctx, job)
				svc.EventBroker.Publish("job.completed", job)
				continue
			}

			// TODO : Run FFMPEG and convert to
			// adaptive Bitrate streamable
			// Content

			key := uuid.New().String()
			videoArtifact := &VideoArtifact{
				ID:             uuid.New(),
				NodeID:         &job.NodeID,
				Key:            &key,
				SizeBytes:      node.SizeBytes,
				LastAccessedAt: time.Now(),
				Metadata:       vm,
			}

			err = svc.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
				if err := tx.Create(videoArtifact).Error; err != nil {
					return err
				}
				if err := tx.Delete(&VideoProcessingJob{NodeID: job.NodeID}).Error; err != nil {
					return err
				}
				return nil
			})

			if err != nil {
				log.Println("error in video worker(video artifact creation): ", err)
			}
			svc.EventBroker.Publish("job.completed", job)
		}
	}
}
