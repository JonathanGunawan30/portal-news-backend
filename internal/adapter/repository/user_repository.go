package repository

import (
	"context"
	"errors"

	"github.com/JonathanGunawan30/portal-news-backend.git/internal/core/domain/entity"
	errors2 "github.com/JonathanGunawan30/portal-news-backend.git/internal/core/domain/errors"
	"github.com/JonathanGunawan30/portal-news-backend.git/internal/core/domain/model"
	"github.com/gofiber/fiber/v2/log"
	"gorm.io/gorm"
)

type UserRepository interface {
	UpdatePassword(ctx context.Context, id int64, password string) error
	GetUserByID(ctx context.Context, id int64) (*entity.UserEntity, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

func (repository *userRepository) UpdatePassword(ctx context.Context, id int64, password string) error {
	result := repository.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Update("password", password)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (repository *userRepository) GetUserByID(ctx context.Context, id int64) (*entity.UserEntity, error) {
	var modelUser model.User
	result := repository.db.WithContext(ctx).Take(&modelUser, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors2.ErrorUserNotFound
		}
		code := "[REPOSITORY] GetUserByID - 1"
		log.Errorw(code, result.Error)
		return nil, result.Error
	}

	return &entity.UserEntity{
		ID:       id,
		Name:     modelUser.Name,
		Email:    modelUser.Email,
		Password: modelUser.Password,
	}, nil
}
