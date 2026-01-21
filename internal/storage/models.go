package storage

import (
	"time"
)

type NodeType string

const (
	NodeTypeFile      NodeType = "file"
	NodeTypeDirectory NodeType = "directory"
)

type Node struct {
	ID               int64     `json:"id" db:"id"`
	OwnerID          int64     `json:"owner_id" db:"owner_id"`
	ParentID         *int64    `json:"parent_id" db:"parent_id"`
	Name             string    `json:"name" db:"name"`
	Type             NodeType  `json:"node_type" db:"node_type"`
	ObjectStorageKey *string   `json:"object_storage_key,omitempty" db:"object_storage_key"` // Only for files
	SHA256Hash       *string   `json:"sha256_hash,omitempty" db:"sha256_hash"`               // Only for files
	SizeBytes        *int64    `json:"size_bytes,omitempty" db:"size_bytes"`                 // Only for files
	MimeType         *string   `json:"mime_type,omitempty" db:"mime_type"`                   // Only for files
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

type Subtitle struct {
	Language   string `json:"language"`
	Label      string `json:"label"`
	StorageKey string `json:"storage_key"`
}

type VideoMetadata struct {
	NodeID          int64      `json:"node_id" db:"node_id"` // Links directly to Node
	Quality         string     `json:"quality,omitempty" db:"quality"`
	BitrateKBPS     int        `json:"bitrate_kbps,omitempty" db:"bitrate_kbps"`
	DurationSeconds int        `json:"duration_seconds,omitempty" db:"duration_seconds"`
	Codec           string     `json:"codec,omitempty" db:"codec"`
	Subtitles       []Subtitle `json:"subtitles,omitempty" db:"subtitles"`
}

type PhotoMetadata struct {
	NodeID        int64      `json:"node_id" db:"node_id"` // Links directly to Node
	Width         int        `json:"width,omitempty" db:"width"`
	Height        int        `json:"height,omitempty" db:"height"`
	CameraMake    string     `json:"camera_make,omitempty" db:"camera_make"`
	CameraModel   string     `json:"camera_model,omitempty" db:"camera_model"`
	ExposureTime  string     `json:"exposure_time,omitempty" db:"exposure_time"`
	ISO           int        `json:"iso,omitempty" db:"iso"`
	TakenAt       *time.Time `json:"taken_at,omitempty" db:"taken_at"` // Pointer for omitempty on nil values
}

type MusicMetadata struct {
	NodeID          int64  `json:"node_id" db:"node_id"` // Links directly to Node
	Title           string `json:"title,omitempty" db:"title"`
	Artist          string `json:"artist,omitempty" db:"artist"`
	Album           string `json:"album,omitempty" db:"album"`
	TrackNumber     int    `json:"track_number,omitempty" db:"track_number"`
	DurationSeconds int    `json:"duration_seconds,omitempty" db:"duration_seconds"`
	Genre           string `json:"genre,omitempty" db:"genre"`
}

type PermissionRole string

const (
	RoleEditor PermissionRole = "editor"
	RoleViewer PermissionRole = "viewer"
)

type NodePermission struct {
	ID        int64          `json:"id" db:"id"`
	NodeID    int64          `json:"node_id" db:"node_id"`
	UserID    int64          `json:"user_id" db:"user_id"`
	Role      PermissionRole `json:"role" db:"role"`
	GrantedAt time.Time      `json:"granted_at" db:"granted_at"`
}