package cleanup

import (
	"context"
	"log"
	"time"

	"github.com/zhitoo/hls_converter/internal/models"
	"github.com/zhitoo/hls_converter/internal/storage"
	"github.com/zhitoo/hls_converter/internal/task"
)

const taskTTL = 24 * time.Hour

type Cleaner struct {
	taskRepo task.Repository
	storage  storage.Storage
}

func New(taskRepo task.Repository, stor storage.Storage) *Cleaner {
	return &Cleaner{taskRepo: taskRepo, storage: stor}
}

// Run checks for expired tasks every interval and deletes them.
func (c *Cleaner) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.sweep()
		}
	}
}

// DrainStale deletes all tasks that were left in Pending or Processing state,
// which can happen when the server is restarted mid-conversion.
func (c *Cleaner) DrainStale() {
	tasks, err := c.taskRepo.ListAll()
	if err != nil {
		log.Printf("cleanup: listing tasks on startup: %v", err)
		return
	}

	for _, t := range tasks {
		if t.Status != models.StatusPending && t.Status != models.StatusProcessing {
			continue
		}
		if err := c.storage.DeleteTaskFiles(t.UserID, t.TaskID); err != nil {
			log.Printf("cleanup: deleting files for stale task %s: %v", t.TaskID, err)
		}
		if err := c.taskRepo.Delete(t.TaskID); err != nil {
			log.Printf("cleanup: deleting stale task record %s: %v", t.TaskID, err)
			continue
		}
		log.Printf("cleanup: removed stale task %s (status: %s)", t.TaskID, t.Status)
	}
}

func (c *Cleaner) sweep() {
	tasks, err := c.taskRepo.ListAll()
	if err != nil {
		log.Printf("cleanup: listing tasks: %v", err)
		return
	}

	cutoff := time.Now().Add(-taskTTL)
	for _, t := range tasks {
		if t.CreatedAt.After(cutoff) {
			continue
		}
		if err := c.storage.DeleteTaskFiles(t.UserID, t.TaskID); err != nil {
			log.Printf("cleanup: deleting files for task %s: %v", t.TaskID, err)
			continue
		}
		if err := c.taskRepo.Delete(t.TaskID); err != nil {
			log.Printf("cleanup: deleting task record %s: %v", t.TaskID, err)
			continue
		}
		log.Printf("cleanup: deleted task %s (created %s)", t.TaskID, t.CreatedAt.Format(time.RFC3339))
	}
}
