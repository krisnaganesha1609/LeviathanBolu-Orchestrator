package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/domain"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/dto"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/repositories"
)

type UserSettingsService interface {
	SetDefaultUserSettings(ctx context.Context, userID uuid.UUID) error
	GetUserSettings(ctx context.Context, userID uuid.UUID) (*domain.UserSettings, error)
	UpdateUserSettings(ctx context.Context, userID uuid.UUID, settings *dto.UpdateUserSettingsRequest) error
}

type UserSettingsServiceImpl struct {
	UserSettingsRepository repositories.UserSettingsRepository
}

func InitUserSettingsService(userSettingsRepository repositories.UserSettingsRepository) UserSettingsService {
	return &UserSettingsServiceImpl{
		UserSettingsRepository: userSettingsRepository,
	}
}

// GetUserSettings implements [UserSettingsService].
func (u *UserSettingsServiceImpl) GetUserSettings(ctx context.Context, userID uuid.UUID) (*domain.UserSettings, error) {
	return u.UserSettingsRepository.GetByUserID(ctx, userID)
}

// SetDefaultUserSettings implements [UserSettingsService].
func (u *UserSettingsServiceImpl) SetDefaultUserSettings(ctx context.Context, userID uuid.UUID) error {
	return u.UserSettingsRepository.Create(ctx, &domain.UserSettings{
		UserID:        userID,
		AssistantName: "LeviathanBolu",
		WakeWord:      []string{"Bolu", "Hey Bolu", "Ok Bolu", "Leviathan", "Rise Leviathan"},
		Language:      "en",
		PreferredLLM:  "none",
		PreferredTTS:  "none",
		Theme:         "dark",
	})
}

// UpdateUserSettings implements [UserSettingsService].
func (u *UserSettingsServiceImpl) UpdateUserSettings(ctx context.Context, userID uuid.UUID, settings *dto.UpdateUserSettingsRequest) error {
	existingSettings, err := u.UserSettingsRepository.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}
	existingSettings.AssistantName = settings.AssistantName
	existingSettings.WakeWord = settings.WakeWord
	existingSettings.Language = settings.Language
	existingSettings.PreferredLLM = settings.PreferredLLM
	existingSettings.PreferredTTS = settings.PreferredTTS
	existingSettings.Theme = settings.Theme
	return u.UserSettingsRepository.Update(ctx, existingSettings)
}
