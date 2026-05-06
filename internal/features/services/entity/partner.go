package entity

import (
	"time"

	"github.com/google/uuid"
)

type Partner struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	ServiceID uuid.UUID
	IsActive  bool
	Rating    float64
	CreatedAt time.Time
}
