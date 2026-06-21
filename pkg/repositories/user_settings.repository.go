package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/domain"
	"gorm.io/gorm"
)

type UserSettingsRepository interface {
	Create(ctx context.Context, settings *domain.UserSettings) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.UserSettings, error)
	Update(ctx context.Context, settings *domain.UserSettings) error
	Delete(ctx context.Context, userID uuid.UUID) error
}

type UserSettingsRepositoryImpl struct {
	db *gorm.DB
}

func InitUserSettingsRepository(db *gorm.DB) UserSettingsRepository {
	return &UserSettingsRepositoryImpl{db: db}
}

// Create implements [UserSettingsRepository].
func (u *UserSettingsRepositoryImpl) Create(ctx context.Context, settings *domain.UserSettings) error {
	return u.db.WithContext(ctx).Create(settings).Error
}

// Delete implements [UserSettingsRepository].
func (u *UserSettingsRepositoryImpl) Delete(ctx context.Context, userID uuid.UUID) error {
	return u.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&domain.UserSettings{}).Error
}

// GetByUserID implements [UserSettingsRepository].
func (u *UserSettingsRepositoryImpl) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.UserSettings, error) {
	var settings domain.UserSettings
	err := u.db.WithContext(ctx).Where("user_id = ?", userID).First(&settings).Error
	if err != nil {
		return nil, err
	}
	return &settings, nil
}

// Update implements [UserSettingsRepository].
func (u *UserSettingsRepositoryImpl) Update(ctx context.Context, settings *domain.UserSettings) error {
	return u.db.WithContext(ctx).Save(settings).Error
}
