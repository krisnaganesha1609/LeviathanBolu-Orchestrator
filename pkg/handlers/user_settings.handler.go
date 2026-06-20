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
	uuid, err := uuid.Parse(userID)
	if err != nil {
		return utils.ResponseBadRequest(c, []utils.ValidationError{
			{
				Field:   "user_id",
				Message: "Invalid user ID format",
			},
		})
	}
	settings, err := u.UserSettingsService.GetUserSettings(c, uuid)
	if err != nil {
		return err
	}
	return utils.ResponseOK(c, "User settings retrieved successfully", settings)
}

// UpdateUserSettings implements [UserSettingsHandler].
func (u *UserSettingsHandlerImpl) UpdateUserSettings(c fiber.Ctx) error {
	userID := c.Params("user_id")
	uuid, err := uuid.Parse(userID)
	if err != nil {
		return utils.ResponseBadRequest(c, []utils.ValidationError{
			{
				Field:   "user_id",
				Message: "Invalid user ID format",
			},
		})
	}
	var req *dto.UpdateUserSettingsRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	err = u.UserSettingsService.UpdateUserSettings(c, uuid, req)
	if err != nil {
		return err
	}
	return utils.ResponseOK(c, "User settings updated successfully", nil)
}
