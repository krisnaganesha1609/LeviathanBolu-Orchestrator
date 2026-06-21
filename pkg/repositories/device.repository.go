package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/domain"
	"gorm.io/gorm"
)

type DeviceRepository interface {
	Create(ctx context.Context, device *domain.Device) error
	GetAll(ctx context.Context) ([]*domain.Device, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Device, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Device, error)
	Update(ctx context.Context, device *domain.Device) error
	UpdateLastSeenAt(ctx context.Context, id uuid.UUID, lastSeenAt time.Time) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type DeviceRepositoryImpl struct {
	db *gorm.DB
}

func InitDeviceRepository(db *gorm.DB) DeviceRepository {
	return &DeviceRepositoryImpl{db: db}
}

// IMPLEMENTATION

// Create implements [DeviceRepository].
func (d *DeviceRepositoryImpl) Create(ctx context.Context, device *domain.Device) error {
	return d.db.WithContext(ctx).Create(device).Error
}

// Delete implements [DeviceRepository].
func (d *DeviceRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return d.db.WithContext(ctx).Where("id = ?", id).Delete(&domain.Device{}).Error
}

// GetByID implements [DeviceRepository].
func (d *DeviceRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*domain.Device, error) {
	var device domain.Device
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&device).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

func (d *DeviceRepositoryImpl) GetAll(ctx context.Context) ([]*domain.Device, error) {
	var devices []*domain.Device
	err := d.db.WithContext(ctx).Find(&devices).Error
	if err != nil {
		return nil, err
	}
	return devices, nil
}

// GetByUserID implements [DeviceRepository].
func (d *DeviceRepositoryImpl) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Device, error) {
	var devices []*domain.Device
	err := d.db.WithContext(ctx).Where("user_id = ?", userID).Find(&devices).Error
	if err != nil {
		return nil, err
	}
	return devices, nil
}

// Update implements [DeviceRepository].
func (d *DeviceRepositoryImpl) Update(ctx context.Context, device *domain.Device) error {
	return d.db.WithContext(ctx).Save(device).Error
}

// UpdateLastSeenAt updates only the last_seen_at column, leaving every
// other column untouched. This exists because a naive Save() on a
// partially-populated struct would overwrite UserID/DeviceName/Platform
// with zero values.
func (d *DeviceRepositoryImpl) UpdateLastSeenAt(ctx context.Context, id uuid.UUID, lastSeenAt time.Time) error {
	return d.db.WithContext(ctx).
		Model(&domain.Device{}).
		Where("id = ?", id).
		Update("last_seen_at", lastSeenAt).Error
}
