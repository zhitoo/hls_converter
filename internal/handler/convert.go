package handler

import (
	"encoding/json"
	"net/http"

	"github.com/zhitoo/hls_converter/internal/auth"
	"github.com/zhitoo/hls_converter/internal/models"
)

func (s *Server) handleConvert(w http.ResponseWriter, r *http.Request) {
	var cfg models.HLSConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if cfg.VideoURL == "" {
		writeError(w, http.StatusBadRequest, "video_url is required")
		return
	}

	user := auth.UserFromContext(r.Context())

	t, err := s.taskMgr.Create(user.UserID, cfg, user.MaxConcurrentTasks)
	if err != nil {
		writeError(w, http.StatusTooManyRequests, err.Error())
		return
	}

	if err := s.queue.Enqueue(t); err != nil {
		writeError(w, http.StatusServiceUnavailable, err.Error())
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]string{
		"task_id": t.TaskID,
		"message": "task created and queued successfully",
	})
}
