package dto

import "github.com/google/uuid"

type CreateUserRequest struct {
	Email string `json:"email" validate:"required,email"`
	Name  string `json:"name" validate:"required"`
}

type UpdateUserRequest struct {
	ID    uuid.UUID `json:"id" validate:"required"`
	Email string    `json:"email" validate:"omitempty,email"`
	Name  string    `json:"name"`
}

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	UpdatedAt string    `json:"updated_at"`
}
