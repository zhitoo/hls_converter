package tasklog

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Logger writes timestamped messages to a task-specific log file and satisfies io.Writer
// so it can be used directly as FFmpeg's stderr sink.
type Logger interface {
	Write(p []byte) (int, error)
	Log(format string, args ...any)
	Close() error
}

// Factory creates loggers for individual tasks.
type Factory interface {
	New(taskID string) (Logger, error)
}

type fileLogger struct {
	mu   sync.Mutex
	file *os.File
}

func (l *fileLogger) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.file.Write(p)
}

func (l *fileLogger) Log(format string, args ...any) {
	msg := fmt.Sprintf("[%s] %s\n", time.Now().Format(time.RFC3339), fmt.Sprintf(format, args...))
	l.mu.Lock()
	defer l.mu.Unlock()
	_, _ = l.file.WriteString(msg)
}

func (l *fileLogger) Close() error {
	return l.file.Close()
}

type fileFactory struct {
	logsDir string
}

func NewFactory(logsDir string) Factory {
	return &fileFactory{logsDir: logsDir}
}

func (f *fileFactory) New(taskID string) (Logger, error) {
	if err := os.MkdirAll(f.logsDir, 0o755); err != nil {
		return nil, err
	}
	path := filepath.Join(f.logsDir, taskID+".log")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}
	return &fileLogger{file: file}, nil
}
