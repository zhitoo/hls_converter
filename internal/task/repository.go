package task

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/zhitoo/hls_converter/internal/models"
)

// Repository persists and retrieves task state on the filesystem.
type Repository interface {
	Save(task *models.Task) error
	FindByID(taskID string) (*models.Task, error)
	CountActiveByUserID(userID string) (int, error)
}

type fileRepository struct {
	mu       sync.RWMutex
	tasksDir string
}

func NewFileRepository(tasksDir string) Repository {
	return &fileRepository{tasksDir: tasksDir}
}

func (r *fileRepository) path(taskID string) string {
	return filepath.Join(r.tasksDir, taskID+".json")
}

func (r *fileRepository) Save(task *models.Task) error {
	data, err := json.MarshalIndent(task, "", "  ")
	if err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	tmp := r.path(task.TaskID) + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, r.path(task.TaskID))
}

func (r *fileRepository) FindByID(taskID string) (*models.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	data, err := os.ReadFile(r.path(taskID))
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}
	if err != nil {
		return nil, err
	}

	var t models.Task
	if err := json.Unmarshal(data, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *fileRepository) CountActiveByUserID(userID string) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entries, err := os.ReadDir(r.tasksDir)
	if os.IsNotExist(err) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	count := 0
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(r.tasksDir, e.Name()))
		if err != nil {
			continue
		}
		var t models.Task
		if err := json.Unmarshal(data, &t); err != nil {
			continue
		}
		if t.UserID == userID && (t.Status == models.StatusPending || t.Status == models.StatusProcessing) {
			count++
		}
	}
	return count, nil
}
