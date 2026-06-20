package handlers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/dto"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/services"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/utils"
)

type DeviceHandler interface {
	RegisterDevice(c fiber.Ctx) error
	GetUserDevices(c fiber.Ctx) error
}

type DeviceHandlerImpl struct {
	DeviceService services.DeviceService
}

func InitDeviceHandler(deviceService services.DeviceService) DeviceHandler {
	return &DeviceHandlerImpl{
		DeviceService: deviceService,
	}
}

// GetUserDevices implements [DeviceHandler].
func (d *DeviceHandlerImpl) GetUserDevices(c fiber.Ctx) error {
	userID := c.Params("user_id")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "user ID is required",
		})
	}
	uuid, err := uuid.Parse(userID)
	if err != nil {
		return utils.ResponseBadRequest(c, []utils.ValidationError{
			{
				Field:   "id",
				Message: "Invalid user ID format",
			},
		})
	}
	devices, err := d.DeviceService.GetDevicesByUserID(c, uuid)
	if err != nil {
		return err
	}
	deviceResponses := make([]dto.DeviceResponse, len(devices))
	for i, device := range devices {
		deviceResponses[i] = dto.DeviceResponse{
			ID:         device.ID,
			UserID:     device.UserID,
			DeviceName: device.DeviceName,
			Platform:   device.Platform,
			LastSeenAt: device.LastSeenAt.Format("2006-01-02 15:04:05"),
		}
	}
	return utils.ResponseOK(c, "Devices retrieved successfully", deviceResponses)
}

// RegisterDevice implements [DeviceHandler].
func (d *DeviceHandlerImpl) RegisterDevice(c fiber.Ctx) error {
	var req dto.CreateDeviceRequest
	if err := c.Bind().Body(&req); err != nil {
		return utils.ResponseBadRequest(c, []string{"invalid request body"})
	}

	if errs := utils.ValidateStruct(req); len(errs) > 0 {
		return utils.ResponseBadRequest(c, errs)
	}

	if err := d.DeviceService.CreateDevice(c, req); err != nil {
		return err
	}

	return utils.ResponseCreated(c, "Device registered successfully")
}
