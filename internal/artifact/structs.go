package artifact

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/sirkartik/cloud_drive_2.0/internal/events"
	"github.com/sirkartik/cloud_drive_2.0/internal/storage"
	"gorm.io/gorm"
)

type Service struct {
	DB                   *gorm.DB
	MaxWorkers           uint8
	StorageSvc           *storage.Service
	ProcessingQueue      chan *events.Job
	NewJobEventBroker    *events.Broker[*events.Job]
	ProgressUpdateBroker *events.Broker[*events.JobProgress]
	VideoEncoder         string
}

// Internal struct to match ffprobe's output schema
type FFprobeOutput struct {
	Streams []struct {
		CodecName string `json:"codec_name"`
		Width     int    `json:"width"`
		Height    int    `json:"height"`
		CodecType string `json:"codec_type"` // "video" or "audio"
	} `json:"streams"`
	Format struct {
		Duration float64 `json:"duration,string"`
		Bitrate  uint32  `json:"bit_rate,string"`
	} `json:"format"`
}

type VideoMetadata struct {
	Duration float64 `json:"duration_seconds"`

	Width  int `json:"width"`
	Height int `json:"height"`

	Codec string `json:"codec"` // h264, hevc, vp9, etc.

	Bitrate  uint32 `json:"bit_rate"`
	HasAudio bool   `json:"has_audio"`

	Error *string `json:"error,omitempty"`
}

func (v VideoMetadata) Value() (driver.Value, error) {
	return json.Marshal(v)
}

func (v *VideoMetadata) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, v)
}
