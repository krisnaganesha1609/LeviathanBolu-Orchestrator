package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Device struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey"`

	UserID uuid.UUID `gorm:"type:uuid;index;not null"`

	DeviceName string `gorm:"not null"`

	Platform string `gorm:"not null"`

	LastSeenAt time.Time

	User User `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// BeforeCreate ensures every Device gets a UUID. See User.BeforeCreate for
// why this can't be left to a `default:gen_random_uuid()` tag alone.
func (d *Device) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}
