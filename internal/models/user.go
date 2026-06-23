package models

type User struct {
	UserID             string `json:"user_id"`
	APIKey             string `json:"api_key"`
	MaxConcurrentTasks int    `json:"max_concurrent_tasks"`
}
