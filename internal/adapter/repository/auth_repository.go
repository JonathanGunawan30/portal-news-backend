package repository

import (
	"context"

	"github.com/JonathanGunawan30/portal-news-backend.git/internal/core/domain/entity"
	"github.com/JonathanGunawan30/portal-news-backend.git/internal/core/domain/model"
	"github.com/gofiber/fiber/v2/log"
	"gorm.io/gorm"
)

type AuthRepository interface {
	GetUserByEmail(ctx context.Context, request entity.LoginRequest) (*entity.UserEntity, error)
}

type authRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) AuthRepository {
	return &authRepository{db: db}
}

func (repository *authRepository) GetUserByEmail(ctx context.Context, request entity.LoginRequest) (*entity.UserEntity, error) {
	var modelUser model.User

	err := repository.db.WithContext(ctx).Where("email = ?", request.Email).First(&modelUser).Error
	if err != nil {
		code := "[REPOSITORY] GetUserByEmail - 1"
		log.Errorw(code, err)
		return nil, err
	}

	response := entity.UserEntity{
		ID:       modelUser.ID,
		Name:     modelUser.Name,
		Email:    modelUser.Email,
		Password: modelUser.Password,
	}

	return &response, nil
}
