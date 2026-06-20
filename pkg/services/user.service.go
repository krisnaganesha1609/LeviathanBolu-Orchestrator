package services

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/domain"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/dto"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/repositories"
)

type UserService interface {
	CreateUser(c fiber.Ctx, req dto.CreateUserRequest) error
	GetUserByEmail(c fiber.Ctx, email string) (*domain.User, error)
	GetUserByID(c fiber.Ctx, id uuid.UUID) (*domain.User, error)
	UpdateUser(c fiber.Ctx, req dto.UpdateUserRequest) error
}

type UserServiceImpl struct {
	UserRepositories    repositories.UserRepository
	UserSettingsService UserSettingsService
}

func InitUserService(userRepositories repositories.UserRepository, userSettingsService UserSettingsService) UserService {
	return &UserServiceImpl{
		UserRepositories:    userRepositories,
		UserSettingsService: userSettingsService,
	}
}

// IMPLEMENTATION

// CreateUser implements [UserService].
func (u *UserServiceImpl) CreateUser(c fiber.Ctx, req dto.CreateUserRequest) error {
	uuid := uuid.New()
	if err := u.UserRepositories.Create(c, &domain.User{
		ID:        uuid,
		Email:     req.Email,
		Name:      req.Name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}); err != nil {
		return err
	}

	if err := u.UserSettingsService.SetDefaultUserSettings(c, uuid); err != nil {
		return err
	}

	return nil
}

// GetUserByEmail implements [UserService].
func (u *UserServiceImpl) GetUserByEmail(c fiber.Ctx, email string) (*domain.User, error) {
	return u.UserRepositories.GetByEmail(c, email)
}

// GetUserByID implements [UserService].
func (u *UserServiceImpl) GetUserByID(c fiber.Ctx, id uuid.UUID) (*domain.User, error) {
	return u.UserRepositories.GetByID(c, id)
}

// UpdateUser implements [UserService].
func (u *UserServiceImpl) UpdateUser(c fiber.Ctx, req dto.UpdateUserRequest) error {
	return u.UserRepositories.Update(c, &domain.User{
		ID:        req.ID,
		Email:     req.Email,
		Name:      req.Name,
		UpdatedAt: time.Now(),
	})
}
