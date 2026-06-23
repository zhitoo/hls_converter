package queue

import (
	"fmt"

	"github.com/zhitoo/hls_converter/internal/models"
)

// Queue accepts tasks from HTTP handlers and delivers them to workers.
type Queue struct {
	ch chan *models.Task
}

func New(bufferSize int) *Queue {
	return &Queue{ch: make(chan *models.Task, bufferSize)}
}

// Enqueue adds a task to the queue. Returns an error if the queue is full.
func (q *Queue) Enqueue(task *models.Task) error {
	select {
	case q.ch <- task:
		return nil
	default:
		return fmt.Errorf("queue is full, try again later")
	}
}

// Channel returns the receive end of the queue for workers to consume.
func (q *Queue) Channel() <-chan *models.Task {
	return q.ch
}
