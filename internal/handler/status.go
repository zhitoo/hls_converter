package handler

import (
	"net/http"
	"time"
)

type statusResponse struct {
	TaskID      string    `json:"task_id"`
	Status      string    `json:"status"`
	Progress    int       `json:"progress"`
	CurrentStep string    `json:"current_step"`
	RetryCount  int       `json:"retry_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	taskID := r.PathValue("task_id")

	t, errMsg := s.requireOwned(r, taskID)
	if errMsg != "" {
		status := http.StatusNotFound
		if errMsg == "access denied" || errMsg == "invalid task_id" {
			status = http.StatusForbidden
		}
		writeError(w, status, errMsg)
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{
		TaskID:      t.TaskID,
		Status:      string(t.Status),
		Progress:    t.Progress,
		CurrentStep: t.CurrentStep,
		RetryCount:  t.RetryCount,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	})
}
