package model

import (
	"time"
)

// Favorite представляет избранную сущность пользователя
type Favorite struct {
	ID         string     `json:"id"`
	UserID     string     `json:"userId"`
	User       *User      `json:"user"`
	EntityType EntityType `json:"entityType"`
	EntityID   string     `json:"entityId"`
	EntityName string     `json:"entityName"`
	Notes      *string    `json:"notes,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
}

// Input типы

// CreateFavoriteInput входные данные для создания избранного
type CreateFavoriteInput struct {
	EntityType EntityType `json:"entityType"`
	EntityID   string     `json:"entityId"`
	EntityName string     `json:"entityName"`
	Notes      *string    `json:"notes,omitempty"`
}

// UpdateFavoriteNotesInput входные данные для обновления заметок
type UpdateFavoriteNotesInput struct {
	ID    string  `json:"id"`
	Notes *string `json:"notes,omitempty"`
}
