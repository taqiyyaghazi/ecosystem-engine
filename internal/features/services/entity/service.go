package entity

import (
	"time"

	"github.com/google/uuid"
)

type Service struct {
	ID          uuid.UUID
	Name        string
	Description string
	Price       float64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
