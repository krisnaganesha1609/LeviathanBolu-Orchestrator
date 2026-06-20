package repositories

import (
	"context"

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
	return d.db.Create(device).Error
}

// Delete implements [DeviceRepository].
func (d *DeviceRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return d.db.Delete(&domain.Device{}, id).Error
}

// GetByID implements [DeviceRepository].
func (d *DeviceRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*domain.Device, error) {
	var device domain.Device
	err := d.db.Where("id = ?", id).First(&device).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

func (d *DeviceRepositoryImpl) GetAll(ctx context.Context) ([]*domain.Device, error) {
	var devices []*domain.Device
	err := d.db.Find(&devices).Error
	if err != nil {
		return nil, err
	}
	return devices, nil
}

// GetByUserID implements [DeviceRepository].
func (d *DeviceRepositoryImpl) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Device, error) {
	var devices []*domain.Device
	err := d.db.Where("user_id = ?", userID).Find(&devices).Error
	if err != nil {
		return nil, err
	}
	return devices, nil
}

// Update implements [DeviceRepository].
func (d *DeviceRepositoryImpl) Update(ctx context.Context, device *domain.Device) error {
	return d.db.Save(device).Error
}
