package handlers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/dto"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/services"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/utils"
)

type UserSettingsHandler interface {
	GetUserSettings(c fiber.Ctx) error
	UpdateUserSettings(c fiber.Ctx) error
}

type UserSettingsHandlerImpl struct {
	UserSettingsService services.UserSettingsService
}

func InitUserSettingsHandler(userSettingsService services.UserSettingsService) UserSettingsHandler {
	return &UserSettingsHandlerImpl{
		UserSettingsService: userSettingsService,
	}
}

// GetUserSettings implements [UserSettingsHandler].
func (u *UserSettingsHandlerImpl) GetUserSettings(c fiber.Ctx) error {
	userID := c.Params("user_id")
	id, err := uuid.Parse(userID)
	if err != nil {
		return utils.ResponseBadRequest(c, []utils.ValidationError{
			{
				Field:   "user_id",
				Message: "Invalid user ID format",
			},
		})
	}
	settings, err := u.UserSettingsService.GetUserSettings(c.Context(), id)
	if err != nil {
		return err
	}
	return utils.ResponseOK(c, "User settings retrieved successfully", dto.UserSettingsResponse{
		AssistantName: settings.AssistantName,
		WakeWord:      settings.WakeWord,
		Language:      settings.Language,
		PreferredLLM:  settings.PreferredLLM,
		PreferredTTS:  settings.PreferredTTS,
		Theme:         settings.Theme,
	})
}

// UpdateUserSettings implements [UserSettingsHandler].
func (u *UserSettingsHandlerImpl) UpdateUserSettings(c fiber.Ctx) error {
	userID := c.Params("user_id")
	id, err := uuid.Parse(userID)
	if err != nil {
		return utils.ResponseBadRequest(c, []utils.ValidationError{
			{
				Field:   "user_id",
				Message: "Invalid user ID format",
			},
		})
	}
	var req dto.UpdateUserSettingsRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	if errs := utils.ValidateStruct(req); len(errs) > 0 {
		return utils.ResponseBadRequest(c, errs)
	}
	if err := u.UserSettingsService.UpdateUserSettings(c.Context(), id, &req); err != nil {
		return err
	}
	return utils.ResponseOK(c, "User settings updated successfully", nil)
}
