package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DeviceSession struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey"`

	DeviceID uuid.UUID `gorm:"type:uuid;index;not null"`

	IsActive bool

	ConnectedAt time.Time

	LastPingAt time.Time

	Device Device `gorm:"foreignKey:DeviceID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// BeforeCreate ensures every DeviceSession gets a UUID. See
// User.BeforeCreate for why this can't be left to a `default:` tag alone.
func (s *DeviceSession) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}
