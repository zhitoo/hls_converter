package models

import "time"

type TaskStatus string

const (
	StatusPending    TaskStatus = "Pending"
	StatusProcessing TaskStatus = "Processing"
	StatusCompleted  TaskStatus = "Completed"
	StatusFailed     TaskStatus = "Failed"
)

type Task struct {
	TaskID      string     `json:"task_id"`
	UserID      string     `json:"user_id"`
	Status      TaskStatus `json:"status"`
	Progress    int        `json:"progress"`
	CurrentStep string     `json:"current_step"`
	RetryCount  int        `json:"retry_count"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Config      HLSConfig  `json:"config"`
}
