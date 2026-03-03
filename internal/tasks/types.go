package tasks

import "time"

type TaskType string

const (
	TaskTypeSemanticSync     TaskType = "semantic_sync"
	TaskTypeTocExtract       TaskType = "toc_extract"
	TaskTypeDeleteBook       TaskType = "delete_book"
	TaskTypeUpdateMetadata   TaskType = "update_metadata"
	TaskTypeCheckMissing     TaskType = "check_missing"
	TaskTypeCopyrightExtract TaskType = "copyright_extract"
)

type TaskMode string

const (
	TaskModeFull        TaskMode = "full"
	TaskModeIncremental TaskMode = "incremental"
)

type TaskStatus struct {
	ID        string    `json:"id"`
	Type      TaskType  `json:"type"`
	Mode      TaskMode  `json:"mode"`
	State     string    `json:"state"` // running, completed, error, stopped
	Progress  float64   `json:"progress"`
	Message   string    `json:"message"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Error     string    `json:"error,omitempty"`
}

type Task interface {
	Run() error
	Stop()
	GetStatus() TaskStatus
}
