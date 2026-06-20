package domain

import (
	"time"

	"github.com/google/uuid"
)

type DeviceSession struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`

	DeviceID uuid.UUID `gorm:"type:uuid"`

	IsActive bool

	ConnectedAt time.Time

	LastPingAt time.Time
}
