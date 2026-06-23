package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/domain"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/dto"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/repositories"
)

type DeviceService interface {
	CreateDevice(ctx context.Context, req dto.CreateDeviceRequest) (string, error)
	GetDeviceByID(ctx context.Context, id uuid.UUID) (*domain.Device, error)
	GetDevicesByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Device, error)
	UpdateLastSeenAt(ctx context.Context, id uuid.UUID) error
}

type DeviceServiceImpl struct {
	DeviceRepo repositories.DeviceRepository
}

func InitDeviceService(deviceRepo repositories.DeviceRepository) DeviceService {
	return &DeviceServiceImpl{
		DeviceRepo: deviceRepo,
	}
}

// IMPLEMENTATION

// CreateDevice implements [DeviceService].
func (d *DeviceServiceImpl) CreateDevice(ctx context.Context, req dto.CreateDeviceRequest) (string, error) {
	device := &domain.Device{
		UserID:     req.UserID,
		DeviceName: req.DeviceName,
		Platform:   req.Platform,
		LastSeenAt: time.Now(),
	}
	if id, err := d.DeviceRepo.Create(ctx, device); err != nil {
		return "", err
	} else {
		return id, nil
	}

}

// GetDeviceByID implements [DeviceService].
func (d *DeviceServiceImpl) GetDeviceByID(ctx context.Context, id uuid.UUID) (*domain.Device, error) {
	return d.DeviceRepo.GetByID(ctx, id)
}

// GetDevicesByUserID implements [DeviceService].
func (d *DeviceServiceImpl) GetDevicesByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Device, error) {
	return d.DeviceRepo.GetByUserID(ctx, userID)
}

// UpdateLastSeenAt implements [DeviceService].
//
// Previously this built a near-empty domain.Device{ID, LastSeenAt} and
// called Save() on it, which performs a full-row UPDATE — silently wiping
// UserID/DeviceName/Platform back to zero values on every heartbeat. It
// now updates only the last_seen_at column via the repository.
func (d *DeviceServiceImpl) UpdateLastSeenAt(ctx context.Context, id uuid.UUID) error {
	return d.DeviceRepo.UpdateLastSeenAt(ctx, id, time.Now())
}
