package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	Email string    `gorm:"unique"`
	Name  string    `gorm:"not null"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
