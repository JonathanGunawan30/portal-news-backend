package handler

import (
	"fmt"
	"time"

	"github.com/JonathanGunawan30/portal-news-backend.git/config"
	"github.com/JonathanGunawan30/portal-news-backend.git/internal/adapter/handler/request"
	"github.com/JonathanGunawan30/portal-news-backend.git/internal/adapter/handler/response"
	"github.com/JonathanGunawan30/portal-news-backend.git/internal/core/domain/entity"
	"github.com/JonathanGunawan30/portal-news-backend.git/internal/core/service"
	"github.com/JonathanGunawan30/portal-news-backend.git/lib/auth"
	"github.com/JonathanGunawan30/portal-news-backend.git/lib/recaptcha"
	validation "github.com/JonathanGunawan30/portal-news-backend.git/validator"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

type AuthHandler interface {
	Login(ctx *fiber.Ctx) error
	Logout(ctx *fiber.Ctx) error
}

type authHandler struct {
	authService service.AuthService
	cfg         *config.Config
	jwtAuth     auth.Jwt
}

func NewAuthHandler(authService service.AuthService, cfg *config.Config, jwtAuth auth.Jwt) AuthHandler {
	return &authHandler{
		authService: authService,
		cfg:         cfg,
		jwtAuth:     jwtAuth,
	}
}

func (handler *authHandler) Login(ctx *fiber.Ctx) error {
	loginRequest := request.LoginRequest{}
	if err := ctx.BodyParser(&loginRequest); err != nil {
		code := "[HANDLER] login - 1"
		log.Errorw(code, err)
		return respondError(ctx, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := validation.ValidateStruct(loginRequest); err != nil {
		code := "[HANDLER] login - 2"
		log.Errorw(code, err)
		return respondError(ctx, fiber.StatusBadRequest, err.Error())
	}

	recaptchaResult, err := recaptcha.VerifyToken(ctx.Context(), handler.cfg.App.RecaptchaSecretKey, loginRequest.RecaptchaToken)
	if err != nil {
		code := "[HANDLER] login - recaptcha_error"
		log.Errorw(code, err)
		return respondError(ctx, fiber.StatusInternalServerError, "Error verifying reCAPTCHA")
	}

	if !recaptchaResult.Success || recaptchaResult.Score < 0.5 {
		code := "[HANDLER] login - recaptcha_failed"
		log.Errorw(code, fmt.Errorf("recaptcha failed: success=%v, score=%f", recaptchaResult.Success, recaptchaResult.Score))
		return respondError(ctx, fiber.StatusForbidden, "reCAPTCHA verification failed")
	}

	userLoginRequest := entity.LoginRequest{
		Email:    loginRequest.Email,
		Password: loginRequest.Password,
	}

	result, err := handler.authService.GetUserByEmail(ctx.Context(), userLoginRequest)
	if err != nil {
		code := "[HANDLER] login - 3"
		log.Errorw(code, err)
		return respondError(ctx, fiber.StatusBadRequest, "Invalid email or password")
	}

	setAuthCookie(ctx, result.AccessToken, time.Unix(result.ExpiredAt, 0), isProductionEnv(handler.cfg))

	var successResp response.DefaultSuccessResponse
	successResp.Meta.Status = true
	successResp.Meta.Message = "Login Successful"
	successResp.Data = fiber.Map{
		"expired_at": result.ExpiredAt,
	}

	return ctx.JSON(successResp)
}

func (handler *authHandler) Logout(ctx *fiber.Ctx) error {
	if token := ctx.Cookies("X-API-TOKEN"); token != "" && handler.jwtAuth != nil {
		if err := handler.jwtAuth.RevokeToken(token); err != nil {
			log.Warnw("[HANDLER] logout - revoke_token", err)
		}
	}

	setAuthCookie(ctx, "", time.Now().Add(-time.Hour), isProductionEnv(handler.cfg))

	var successResp response.DefaultSuccessResponse
	successResp.Meta.Status = true
	successResp.Meta.Message = "Logout Successful"
	successResp.Data = nil

	return ctx.JSON(successResp)
}
