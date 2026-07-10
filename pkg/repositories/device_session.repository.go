package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/domain"
	"gorm.io/gorm"
)

type DeviceSessionRepository interface {
	Create(ctx context.Context, session *domain.DeviceSession) (string, error)
	GetAll(ctx context.Context) ([]*domain.DeviceSession, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.DeviceSession, error)
	Update(ctx context.Context, session *domain.DeviceSession) error
	GetActiveSessionByDevice(ctx context.Context, deviceID uuid.UUID) (*domain.DeviceSession, error)
}

type DeviceSessionRepositoryImpl struct {
	db *gorm.DB
}

// GetActiveSessionByDevice implements [DeviceSessionRepository].
func (d *DeviceSessionRepositoryImpl) GetActiveSessionByDevice(ctx context.Context, deviceID uuid.UUID) (*domain.DeviceSession, error) {
	var device domain.DeviceSession
	err := d.db.WithContext(ctx).Where("device_id = ? AND ended_at IS NULL", deviceID).First(&device).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

func InitDeviceSessionRepository(db *gorm.DB) DeviceSessionRepository {
	return &DeviceSessionRepositoryImpl{db: db}
}

// IMPLEMENTATION

// Create implements [DeviceSessionRepository].
func (d *DeviceSessionRepositoryImpl) Create(ctx context.Context, device *domain.DeviceSession) (string, error) {
	err := d.db.WithContext(ctx).Create(device).Error
	if err != nil {
		return "", err
	}
	return device.ID.String(), nil
}

// GetByID implements [DeviceSessionRepository].
func (d *DeviceSessionRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*domain.DeviceSession, error) {
	var device domain.DeviceSession
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&device).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

func (d *DeviceSessionRepositoryImpl) GetAll(ctx context.Context) ([]*domain.DeviceSession, error) {
	var devices []*domain.DeviceSession
	err := d.db.WithContext(ctx).Find(&devices).Error
	if err != nil {
		return nil, err
	}
	return devices, nil
}

// Update implements [DeviceRepository].
func (d *DeviceSessionRepositoryImpl) Update(ctx context.Context, device *domain.DeviceSession) error {
	return d.db.WithContext(ctx).Save(device).Error
}
