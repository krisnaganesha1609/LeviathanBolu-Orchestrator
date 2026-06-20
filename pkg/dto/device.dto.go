package dto

import "github.com/google/uuid"

type CreateDeviceRequest struct {
	UserID     uuid.UUID `json:"user_id" validate:"required,uuid4"`
	DeviceName string    `json:"device_name" validate:"required"`
	Platform   string    `json:"platform" validate:"required"`
}

type DeviceResponse struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	DeviceName string    `json:"device_name"`
	Platform   string    `json:"platform"`
	LastSeenAt string    `json:"last_seen_at"`
}
