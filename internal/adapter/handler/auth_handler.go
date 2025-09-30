package handler

import (
	"os"
	"time"

	"github.com/JonathanGunawan30/portal-news-backend.git/internal/adapter/handler/request"
	"github.com/JonathanGunawan30/portal-news-backend.git/internal/adapter/handler/response"
	"github.com/JonathanGunawan30/portal-news-backend.git/internal/core/domain/entity"
	"github.com/JonathanGunawan30/portal-news-backend.git/internal/core/service"
	validation "github.com/JonathanGunawan30/portal-news-backend.git/validator"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

var errResp response.ErrorResponseDefault

type AuthHandler interface {
	Login(ctx *fiber.Ctx) error
	Logout(ctx *fiber.Ctx) error
}

type authHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) AuthHandler {
	return &authHandler{authService: authService}
}

func (handler *authHandler) Login(ctx *fiber.Ctx) error {
	loginRequest := request.LoginRequest{}
	if err := ctx.BodyParser(&loginRequest); err != nil {
		code := "[HANDLER] login - 1"
		log.Errorw(code, err)
		errResp.Meta.Status = false
		errResp.Meta.Message = err.Error()
		return ctx.Status(fiber.StatusBadRequest).JSON(errResp)
	}

	if err := validation.ValidateStruct(loginRequest); err != nil {
		code := "[HANDLER] login - 2"
		log.Errorw(code, err)
		errResp.Meta.Status = false
		errResp.Meta.Message = err.Error()
		return ctx.Status(fiber.StatusBadRequest).JSON(errResp)
	}

	userLoginRequest := entity.LoginRequest{
		Email:    loginRequest.Email,
		Password: loginRequest.Password,
	}

	result, err := handler.authService.GetUserByEmail(ctx.Context(), userLoginRequest)
	if err != nil {
		code := "[HANDLER] login - 3"
		log.Errorw(code, err)
		errResp.Meta.Status = false
		errResp.Meta.Message = err.Error()
		return ctx.Status(fiber.StatusBadRequest).JSON(errResp)
	}

	sameSite := "Lax"
	secure := false
	if os.Getenv("APP_ENV") == "production" {
		sameSite = "None"
		secure = true
	}

	ctx.Cookie(&fiber.Cookie{
		Name:     "X-API-TOKEN",
		Value:    result.AccessToken,
		Expires:  time.Unix(result.ExpiredAt, 0),
		HTTPOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		Path:     "/",
	})

	var successResp response.DefaultSuccessResponse
	successResp.Meta.Status = true
	successResp.Meta.Message = "Login Successful"
	successResp.Data = fiber.Map{
		"expired_at": result.ExpiredAt,
	}

	return ctx.JSON(successResp)

}

func (handler *authHandler) Logout(ctx *fiber.Ctx) error {
	ctx.Cookie(&fiber.Cookie{
		Name:     "X-API-TOKEN",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		Secure:   false, // ganti true kalau sudah pakai HTTPS
		SameSite: "Lax",
		Path:     "/",
	})

	var successResp response.DefaultSuccessResponse
	successResp.Meta.Status = true
	successResp.Meta.Message = "Logout Successful"
	successResp.Data = nil

	return ctx.JSON(successResp)
}
