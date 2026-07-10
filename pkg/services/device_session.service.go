package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/domain"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/repositories"
)

type DeviceSessionService interface {
	CreateSession(ctx context.Context, deviceID uuid.UUID) (string, error)
	GetActiveSessionByDevice(ctx context.Context, deviceID uuid.UUID) (*domain.DeviceSession, error)
	CloseSession(ctx context.Context, id uuid.UUID) error
	UpdatePing(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.DeviceSession, error)
}

type DeviceSessionServiceImpl struct {
	DeviceSessionRepo repositories.DeviceSessionRepository
}

func InitDeviceSessionService(deviceSessionRepo repositories.DeviceSessionRepository) DeviceSessionService {
	return &DeviceSessionServiceImpl{
		DeviceSessionRepo: deviceSessionRepo,
	}
}

// FindByID implements [DeviceSessionService].
func (d *DeviceSessionServiceImpl) FindByID(ctx context.Context, id uuid.UUID) (*domain.DeviceSession, error) {
	return d.DeviceSessionRepo.GetByID(ctx, id)
}

// UpdatePing implements [DeviceSessionService].
func (d *DeviceSessionServiceImpl) UpdatePing(ctx context.Context, id uuid.UUID) error {
	session, err := d.DeviceSessionRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	session.LastPingAt = time.Now()
	return d.DeviceSessionRepo.Update(ctx, session)
}

// CloseSession implements [DeviceSessionService].
func (d *DeviceSessionServiceImpl) CloseSession(ctx context.Context, id uuid.UUID) error {
	session, err := d.DeviceSessionRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	session.EndedAt = time.Now()
	return d.DeviceSessionRepo.Update(ctx, session)
}

// CreateSession implements [DeviceSessionService].
func (d *DeviceSessionServiceImpl) CreateSession(ctx context.Context, deviceID uuid.UUID) (string, error) {
	session := &domain.DeviceSession{
		DeviceID:   deviceID,
		StartedAt:  time.Now(),
		LastPingAt: time.Now(),
	}
	if id, err := d.DeviceSessionRepo.Create(ctx, session); err != nil {
		return "", err
	} else {
		return id, nil
	}

}

// GetActiveSessionByDevice implements [DeviceSessionService].
func (d *DeviceSessionServiceImpl) GetActiveSessionByDevice(ctx context.Context, deviceID uuid.UUID) (*domain.DeviceSession, error) {
	return d.DeviceSessionRepo.GetActiveSessionByDevice(ctx, deviceID)
}
