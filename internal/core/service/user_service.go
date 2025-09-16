package service

import (
	"context"

	"github.com/JonathanGunawan30/portal-news-backend.git/internal/adapter/repository"
	"github.com/JonathanGunawan30/portal-news-backend.git/internal/core/domain/entity"
	"github.com/JonathanGunawan30/portal-news-backend.git/internal/core/domain/errors"
	"github.com/JonathanGunawan30/portal-news-backend.git/lib/conv"
	"github.com/gofiber/fiber/v2/log"
)

type UserService interface {
	UpdatePassword(ctx context.Context, id int64, currentPassword, newPassword string) error
	GetUserByID(ctx context.Context, id int64) (*entity.UserEntity, error)
}

type userService struct {
	userRepository repository.UserRepository
}

func NewUserService(userRepository repository.UserRepository) UserService {
	return &userService{
		userRepository: userRepository,
	}
}

func (service *userService) UpdatePassword(ctx context.Context, id int64, currentPassword, newPassword string) error {

	user, err := service.userRepository.GetUserByID(ctx, id)
	if err != nil {
		code := "[SERVICE] UpdatePassword - 1"
		log.Errorw(code, err)
		return err
	}

	err = conv.ComparePassword(user.Password, currentPassword)
	if err != nil {
		code := "[SERVICE] UpdatePassword - 2"
		log.Errorw(code, "reason", "invalid current password")
		return errors.ErrorPasswordInvalid
	}

	hashPassword, err := conv.HashPassword(newPassword)
	if err != nil {
		code := "[SERVICE] UpdatePassword - 3"
		log.Errorw(code, err)
		return err
	}

	err = service.userRepository.UpdatePassword(ctx, id, hashPassword)
	if err != nil {
		code := "[SERVICE] UpdatePassword - 2"
		log.Errorw(code, err)
		return err
	}
	return nil

}

func (service *userService) GetUserByID(ctx context.Context, id int64) (*entity.UserEntity, error) {
	user, err := service.userRepository.GetUserByID(ctx, id)
	if err != nil {
		code := "[SERVICE] GetUserByID - 1"
		log.Errorw(code, err)
		return nil, err
	}
	return user, nil
}
