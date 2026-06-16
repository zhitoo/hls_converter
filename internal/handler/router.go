package handler

import "net/http"

// Routes registers all API endpoints and returns the root handler.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("POST /api/v1/convert", s.protected(s.handleConvert))
	mux.Handle("GET /api/v1/status/{task_id}", s.protected(s.handleStatus))
	mux.Handle("GET /api/v1/download/{task_id}", s.protected(s.handleDownload))
	mux.Handle("GET /api/v1/logs/{task_id}", s.protected(s.handleLogs))

	return mux
}
