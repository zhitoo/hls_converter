package handler

import (
	"net/http"
	"os"
)

func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	taskID := r.PathValue("task_id")

	_, errMsg := s.requireOwned(r, taskID)
	if errMsg != "" {
		status := http.StatusNotFound
		if errMsg == "access denied" || errMsg == "invalid task_id" {
			status = http.StatusForbidden
		}
		writeError(w, status, errMsg)
		return
	}

	logPath := s.stor.LogFilePath(taskID)
	data, err := os.ReadFile(logPath)
	if os.IsNotExist(err) {
		writeJSON(w, http.StatusOK, map[string]string{"logs": ""})
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "cannot read log file")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"logs": string(data)})
}
