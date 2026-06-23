package task

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/zhitoo/hls_converter/internal/models"
)

// Manager handles task creation and enforces per-user concurrency limits.
type Manager struct {
	repo Repository
}

func NewManager(repo Repository) *Manager {
	return &Manager{repo: repo}
}

// Create validates the concurrency limit then persists a new Pending task.
func (m *Manager) Create(userID string, cfg models.HLSConfig, maxConcurrent int) (*models.Task, error) {
	active, err := m.repo.CountActiveByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("checking active tasks: %w", err)
	}
	if active >= maxConcurrent {
		return nil, fmt.Errorf("concurrent task limit reached (%d/%d)", active, maxConcurrent)
	}

	cfg.SetDefaults()

	t := &models.Task{
		TaskID:      newUUID(),
		UserID:      userID,
		Status:      models.StatusPending,
		Progress:    0,
		CurrentStep: "Queued",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Config:      cfg,
	}

	if err := m.repo.Save(t); err != nil {
		return nil, fmt.Errorf("saving task: %w", err)
	}
	return t, nil
}

func newUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant bits
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
