package artifact

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"fmt"

	"path/filepath"

	"github.com/google/uuid"
	"github.com/sirkartik/cloud_drive_2.0/internal/events"
	"gorm.io/gorm"
)

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
	videoCh := svc.NewJobEventBroker.Subscribe("video")
	jobCompletedCh := svc.NewJobEventBroker.Subscribe("job.completed")

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
	}
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
				svc.NewJobEventBroker.Publish("job.completed", job)
				continue
			}

			filePath := fmt.Sprintf("videos/%v_%s", WorkerID, node.Name)
			vm, err := svc.ffprobe(ctx, filePath)

			if err != nil {
				log.Println("error in video worker: (ffprobe)", err)
				os.Remove(filePath)
				err = svc.setJobStatusFailed(ctx, job)
				svc.NewJobEventBroker.Publish("job.completed", job)
				continue
			}

			filename := filepath.Base(filePath)
			baseName := strings.TrimSuffix(filename, filepath.Ext(filename))
			outputDir := fmt.Sprintf("videos/%v", baseName)
			err = svc.ffmpeg(ctx, filePath, outputDir, vm.Duration, vm.Height, func(percent float64) {
				// TODO send this progress to the frontend!
				log.Printf("Current progress of worker %v: %v\n", WorkerID, percent)
				svc.ProgressUpdateBroker.Publish("progress.update", &events.JobProgress{
					Job:      *job,
					Progress: percent,
				})
			})

			// TODO 1. Add option to cancel the video
			// conversion and revert the changes

			// TODO 2. Give progress to frontend - Partially done
			// Add web socket to push progress to frontend

			// TODO 3. Generate signed URLs for frontend
			if err != nil {
				log.Println("error in video worker: (ffmpeg)", err)
				err = svc.setJobStatusFailed(ctx, job)
				svc.NewJobEventBroker.Publish("job.completed", job)
				continue
			}

			key := uuid.New().String()

			log.Printf("Worker %v uploading HLS to minio...\n", WorkerID)

			err = svc.StorageSvc.PutHLS(ctx, outputDir, key)

			if err != nil {
				log.Println("error in video worker: (put HLS)", err)
				err = svc.setJobStatusFailed(ctx, job)
				svc.NewJobEventBroker.Publish("job.completed", job)
				os.RemoveAll(outputDir)
				continue
			}
			os.Remove(filePath)
			os.RemoveAll(outputDir)

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
			svc.NewJobEventBroker.Publish("job.completed", job)
		}
	}
}
