package service

import (
	"context"
	"errors"

	"github.com/JonathanGunawan30/portal-news-backend.git/config"
	"github.com/JonathanGunawan30/portal-news-backend.git/internal/adapter/repository"
	"github.com/JonathanGunawan30/portal-news-backend.git/internal/core/domain/entity"
	"github.com/JonathanGunawan30/portal-news-backend.git/lib/auth"
	"github.com/JonathanGunawan30/portal-news-backend.git/lib/conv"
	"github.com/gofiber/fiber/v2/log"
)

type AuthService interface {
	GetUserByEmail(ctx context.Context, request entity.LoginRequest) (*entity.AccessToken, error)
}

type authService struct {
	authRepository repository.AuthRepository
	cfg            *config.Config
	jwtToken       auth.Jwt
}

func NewAuthService(authRepository repository.AuthRepository, cfg *config.Config, jwtToken auth.Jwt) AuthService {
	return &authService{
		authRepository: authRepository,
		cfg:            cfg,
		jwtToken:       jwtToken,
	}
}

func (service *authService) GetUserByEmail(ctx context.Context, request entity.LoginRequest) (*entity.AccessToken, error) {
	result, err := service.authRepository.GetUserByEmail(ctx, request)
	if err != nil {
		code := "[SERIVCE] GetUserByEmail - 1"
		log.Errorw(code, err)
		return nil, err
	}

	if passwordCheckResult := conv.CheckPasswordHash(request.Password, result.Password); !passwordCheckResult {
		code := "[SERIVCE] GetUserByEmail - 2"
		err = errors.New("username or password is wrong")
		log.Errorw(code, err)
		return nil, err
	}

	jwtData := entity.JwtData{
		UserID: result.ID,
	}

	token, expiredAt, err := service.jwtToken.GenerateToken(&jwtData)
	if err != nil {
		code := "[SERIVCE] GetUserByEmail - 3"
		log.Errorw(code, err)
		return nil, err
	}

	response := entity.AccessToken{
		AccessToken: token,
		ExpiredAt:   expiredAt,
	}

	return &response, nil

}
