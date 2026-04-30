package entity

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID
	Username  string
	Password  string
	Role      string
	CreatedAt time.Time
}
