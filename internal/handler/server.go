package handler

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/zhitoo/hls_converter/internal/auth"
	"github.com/zhitoo/hls_converter/internal/models"
	"github.com/zhitoo/hls_converter/internal/queue"
	"github.com/zhitoo/hls_converter/internal/storage"
	"github.com/zhitoo/hls_converter/internal/task"
)

// Server holds all handler dependencies and wires the HTTP routes.
type Server struct {
	taskMgr  *task.Manager
	taskRepo task.Repository
	stor     storage.Storage
	queue    *queue.Queue
	authMW   func(http.Handler) http.Handler
}

func NewServer(
	taskMgr *task.Manager,
	taskRepo task.Repository,
	stor storage.Storage,
	q *queue.Queue,
	authMW func(http.Handler) http.Handler,
) *Server {
	return &Server{
		taskMgr:  taskMgr,
		taskRepo: taskRepo,
		stor:     stor,
		queue:    q,
		authMW:   authMW,
	}
}

var uuidRE = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// requireOwned looks up a task by ID and verifies that the authenticated user owns it.
// Returns (task, "") on success or ("", errorMessage) on failure.
func (s *Server) requireOwned(r *http.Request, taskID string) (*models.Task, string) {
	if !uuidRE.MatchString(taskID) {
		return nil, "invalid task_id"
	}

	t, err := s.taskRepo.FindByID(taskID)
	if err != nil {
		return nil, "task not found"
	}

	user := auth.UserFromContext(r.Context())
	if t.UserID != user.UserID {
		return nil, fmt.Sprintf("access denied")
	}
	return t, ""
}

func (s *Server) protected(h http.HandlerFunc) http.Handler {
	return s.authMW(h)
}
