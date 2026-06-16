package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Storage manages filesystem paths for HLS output and task logs.
type Storage interface {
	HLSOutputDir(userID, taskID string) (string, error)
	LogFilePath(taskID string) string
}

type localStorage struct {
	storageDir string
}

// New creates a Storage rooted at storageDir.
// Layout:
//
//	<storageDir>/users/<userID>/<taskID>/  — HLS output
//	<storageDir>/logs/<taskID>.log         — conversion logs
func New(storageDir string) (Storage, error) {
	abs, err := filepath.Abs(storageDir)
	if err != nil {
		return nil, err
	}
	for _, sub := range []string{"users", "logs"} {
		if err := os.MkdirAll(filepath.Join(abs, sub), 0o755); err != nil {
			return nil, fmt.Errorf("creating %s dir: %w", sub, err)
		}
	}
	return &localStorage{storageDir: abs}, nil
}

func (s *localStorage) HLSOutputDir(userID, taskID string) (string, error) {
	if err := validatePathSegment(userID); err != nil {
		return "", err
	}
	if err := validatePathSegment(taskID); err != nil {
		return "", err
	}
	dir := filepath.Join(s.storageDir, "users", userID, taskID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("creating HLS output dir: %w", err)
	}
	return dir, nil
}

func (s *localStorage) LogFilePath(taskID string) string {
	return filepath.Join(s.storageDir, "logs", taskID+".log")
}

// validatePathSegment rejects IDs containing path traversal characters.
func validatePathSegment(id string) error {
	if strings.Contains(id, "..") || strings.ContainsAny(id, "/\\") {
		return fmt.Errorf("invalid id %q: path traversal detected", id)
	}
	return nil
}
