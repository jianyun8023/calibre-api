package tasks

import "time"

type TaskType string

const (
	TaskTypeQdrantSync TaskType = "qdrant_sync"
	TaskTypeTocExtract TaskType = "toc_extract"
)

type TaskMode string

const (
	TaskModeFull        TaskMode = "full"
	TaskModeIncremental TaskMode = "incremental"
)

type TaskStatus struct {
	ID          string    `json:"id"`
	Type        TaskType  `json:"type"`
	Mode        TaskMode  `json:"mode"`
	State       string    `json:"state"` // running, completed, error, stopped
	Progress    float64   `json:"progress"`
	Message     string    `json:"message"`
	TargetIndex string    `json:"target_index,omitempty"` // For Meilisearch
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Error       string    `json:"error,omitempty"`
}

type Task interface {
	Run() error
	Stop()
	GetStatus() TaskStatus
}
