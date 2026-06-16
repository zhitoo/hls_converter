package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/zhitoo/hls_converter/internal/models"
)

// Repository loads and looks up users.
type Repository interface {
	FindByToken(token string) (*models.User, error)
}

type fileRepository struct {
	mu    sync.RWMutex
	users map[string]*models.User // keyed by APIKey
}

func NewFileRepository(usersFilePath string) (Repository, error) {
	f, err := os.Open(usersFilePath)
	if err != nil {
		return nil, fmt.Errorf("open users file: %w", err)
	}
	defer f.Close()

	var list []*models.User
	if err := json.NewDecoder(f).Decode(&list); err != nil {
		return nil, fmt.Errorf("decode users: %w", err)
	}

	byToken := make(map[string]*models.User, len(list))
	for _, u := range list {
		byToken[u.APIKey] = u
	}

	return &fileRepository{users: byToken}, nil
}

func (r *fileRepository) FindByToken(token string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	u, ok := r.users[token]
	if !ok {
		return nil, fmt.Errorf("invalid token")
	}
	return u, nil
}
