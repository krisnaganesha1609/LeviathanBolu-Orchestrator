package handlers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/dto"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/services"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/utils"
)

type UserHandler interface {
	RegisterUser(c fiber.Ctx) error
	GetUserByEmail(c fiber.Ctx) error
	GetUserByID(c fiber.Ctx) error
	UpdateUser(c fiber.Ctx) error
}

type UserHandlerImpl struct {
	UserService services.UserService
}

func InitUserHandler(userService services.UserService) UserHandler {
	return &UserHandlerImpl{
		UserService: userService,
	}
}

// GetUserByEmail implements [UserHandler].
func (u *UserHandlerImpl) GetUserByEmail(c fiber.Ctx) error {
	userEmail := c.Params("email")
	user, err := u.UserService.GetUserByEmail(c.Context(), userEmail)
	if err != nil {
		return err
	}

	return utils.ResponseOK(c, "User retrieved successfully", dto.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		UpdatedAt: user.UpdatedAt.Format("2006-01-02 15:04:05"),
	})
}

// GetUserByID implements [UserHandler].
func (u *UserHandlerImpl) GetUserByID(c fiber.Ctx) error {
	userID := c.Params("id")
	id, err := uuid.Parse(userID)
	if err != nil {
		return utils.ResponseBadRequest(c, []utils.ValidationError{
			{
				Field:   "id",
				Message: "Invalid user ID format",
			},
		})
	}
	user, err := u.UserService.GetUserByID(c.Context(), id)
	if err != nil {
		return err
	}
	return utils.ResponseOK(c, "User retrieved successfully", dto.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		UpdatedAt: user.UpdatedAt.Format("2006-01-02 15:04:05"),
	})
}

// RegisterUser implements [UserHandler].
func (u *UserHandlerImpl) RegisterUser(c fiber.Ctx) error {
	var req dto.CreateUserRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	if errs := utils.ValidateStruct(req); len(errs) > 0 {
		return utils.ResponseBadRequest(c, errs)
	}
	if err := u.UserService.CreateUser(c.Context(), req); err != nil {
		return err
	}
	return utils.ResponseCreated(c, "user")
}

// UpdateUser implements [UserHandler].
func (u *UserHandlerImpl) UpdateUser(c fiber.Ctx) error {
	userID := c.Params("id")
	id, err := uuid.Parse(userID)
	if err != nil {
		return utils.ResponseBadRequest(c, []utils.ValidationError{
			{
				Field:   "id",
				Message: "Invalid user ID format",
			},
		})
	}

	var req dto.UpdateUserRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	req.ID = id

	if errs := utils.ValidateStruct(req); len(errs) > 0 {
		return utils.ResponseBadRequest(c, errs)
	}

	if err := u.UserService.UpdateUser(c.Context(), req); err != nil {
		return err
	}
	return utils.ResponseEdited(c, "user", "edited")
}
