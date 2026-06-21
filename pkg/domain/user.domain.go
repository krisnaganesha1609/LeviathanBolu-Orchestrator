package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	Email string    `gorm:"unique;not null"`
	Name  string    `gorm:"not null"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

// BeforeCreate ensures every User gets a UUID even if the caller forgot to
// set one. Since GORM 1.25, a zero-value uuid.UUID is no longer treated as
// "blank", so it is sent to Postgres as-is instead of falling back to a
// DB-side default — relying on a `default:` tag alone is not enough.
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
