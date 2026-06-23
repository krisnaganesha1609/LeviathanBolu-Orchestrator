package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/domain"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/dto"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/repositories"
)

type UserService interface {
	CreateUser(ctx context.Context, req dto.CreateUserRequest) (uuid.UUID, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	UpdateUser(ctx context.Context, req dto.UpdateUserRequest) error
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
func (u *UserServiceImpl) CreateUser(ctx context.Context, req dto.CreateUserRequest) (uuid.UUID, error) {
	newUser := &domain.User{
		ID:    uuid.New(),
		Email: req.Email,
		Name:  req.Name,
	}
	if err := u.UserRepositories.Create(ctx, newUser); err != nil {
		return uuid.Nil, err
	}

	if err := u.UserSettingsService.SetDefaultUserSettings(ctx, newUser.ID); err != nil {
		return uuid.Nil, err
	}

	return newUser.ID, nil
}

// GetUserByEmail implements [UserService].
func (u *UserServiceImpl) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	return u.UserRepositories.GetByEmail(ctx, email)
}

// GetUserByID implements [UserService].
func (u *UserServiceImpl) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return u.UserRepositories.GetByID(ctx, id)
}

// UpdateUser implements [UserService].
//
// This performs a partial update: only non-empty fields in the request
// overwrite the stored record. The previous implementation built a brand
// new domain.User from the request and Save()'d it directly, which would
// silently null out CreatedAt and any field the caller didn't (or
// couldn't, since DTO fields aren't pointers) send.
func (u *UserServiceImpl) UpdateUser(ctx context.Context, req dto.UpdateUserRequest) error {
	existing, err := u.UserRepositories.GetByID(ctx, req.ID)
	if err != nil {
		return err
	}

	if req.Email != "" {
		existing.Email = req.Email
	}
	if req.Name != "" {
		existing.Name = req.Name
	}
	existing.UpdatedAt = time.Now()

	return u.UserRepositories.Update(ctx, existing)
}
