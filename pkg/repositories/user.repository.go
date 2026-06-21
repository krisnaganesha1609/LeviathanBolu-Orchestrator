package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/domain"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type UserRepositoryImpl struct {
	db *gorm.DB
}

func InitUserRepository(db *gorm.DB) UserRepository {
	return &UserRepositoryImpl{db: db}
}

// IMPLEMENTATION

func (r *UserRepositoryImpl) Create(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// GetByID looks up a user by primary key.
//
// NOTE: gorm.First(&user, id) only auto-detects the primary key clause for
// numeric/string-shaped conds; with a uuid.UUID value it builds an invalid
// "WHERE <uuid-literal>" clause instead of "WHERE id = '<uuid>'". An
// explicit Where(...) avoids that pitfall.
func (r *UserRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepositoryImpl) Update(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *UserRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&domain.User{}).Error
}

func (r *UserRepositoryImpl) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
