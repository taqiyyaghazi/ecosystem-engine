package entity

import (
	"time"

	"github.com/google/uuid"
)

type Task struct {
	ID           uuid.UUID
	TaskType     string
	Payload      []byte // Stored as JSONB in DB
	Status       string
	ErrorMessage *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
