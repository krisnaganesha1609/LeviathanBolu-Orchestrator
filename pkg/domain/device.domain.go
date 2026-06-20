package domain

import (
	"time"

	"github.com/google/uuid"
)

type Device struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`

	UserID uuid.UUID `gorm:"index,type:uuid"`

	DeviceName string

	Platform string

	LastSeenAt time.Time

	User User `gorm:"foreignKey:UserID"`
}
