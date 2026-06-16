package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/zhitoo/hls_converter/internal/models"
)

type qualityInfo struct {
	Height   int    `json:"height"`
	Label    string `json:"label"`
	Playlist string `json:"playlist"`
}

type statusResponse struct {
	TaskID         string        `json:"task_id"`
	Status         string        `json:"status"`
	Progress       int           `json:"progress"`
	CurrentStep    string        `json:"current_step"`
	RetryCount     int           `json:"retry_count"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
	Qualities      []qualityInfo `json:"qualities,omitempty"`
	MasterPlaylist string        `json:"master_playlist,omitempty"`
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

	resp := statusResponse{
		TaskID:      t.TaskID,
		Status:      string(t.Status),
		Progress:    t.Progress,
		CurrentStep: t.CurrentStep,
		RetryCount:  t.RetryCount,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}

	if t.Status == models.StatusCompleted {
		resp.Qualities, resp.MasterPlaylist = buildQualities(t.Config.Resolutions)
	}

	writeJSON(w, http.StatusOK, resp)
}

func buildQualities(resolutions []int) ([]qualityInfo, string) {
	if len(resolutions) == 0 {
		return []qualityInfo{{Height: 0, Label: "original", Playlist: "output.m3u8"}}, ""
	}

	qualities := make([]qualityInfo, len(resolutions))
	for i, h := range resolutions {
		qualities[i] = qualityInfo{
			Height:   h,
			Label:    fmt.Sprintf("%dp", h),
			Playlist: fmt.Sprintf("%dp/output.m3u8", h),
		}
	}

	master := ""
	if len(resolutions) > 1 {
		master = "master.m3u8"
	}
	return qualities, master
}
