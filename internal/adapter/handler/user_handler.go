package handler

import (
	"errors"
	"time"

	"github.com/JonathanGunawan30/portal-news-backend.git/config"
	"github.com/JonathanGunawan30/portal-news-backend.git/internal/adapter/handler/request"
	"github.com/JonathanGunawan30/portal-news-backend.git/internal/adapter/handler/response"
	"github.com/JonathanGunawan30/portal-news-backend.git/internal/core/domain/entity"
	errors2 "github.com/JonathanGunawan30/portal-news-backend.git/internal/core/domain/errors"
	"github.com/JonathanGunawan30/portal-news-backend.git/internal/core/service"
	"github.com/JonathanGunawan30/portal-news-backend.git/lib/auth"
	validation "github.com/JonathanGunawan30/portal-news-backend.git/validator"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"gorm.io/gorm"
)

type UserHandler interface {
	UpdatePassword(ctx *fiber.Ctx) error
	GetUserByID(ctx *fiber.Ctx) error
}

type userHandler struct {
	userService service.UserService
	cfg         *config.Config
	jwtAuth     auth.Jwt
}

func NewUserHandler(userService service.UserService, cfg *config.Config, jwtAuth auth.Jwt) UserHandler {
	return &userHandler{
		userService: userService,
		cfg:         cfg,
		jwtAuth:     jwtAuth,
	}
}

func (handler *userHandler) UpdatePassword(ctx *fiber.Ctx) error {
	userID, ok := getAuthenticatedUserID(ctx)
	if !ok {
		code := "[HANDLER] UpdatePassword - 1"
		log.Errorw(code, "reason", "unauthorized")
		return respondError(ctx, fiber.StatusUnauthorized, "Unauthorized")
	}

	var req request.UpdatePasswordRequest
	if err := ctx.BodyParser(&req); err != nil {
		code := "[HANDLER] UpdatePassword - 2"
		log.Errorw(code, err)
		return respondError(ctx, fiber.StatusBadRequest, "Failed to parse request body")
	}
	if err := validation.ValidateStruct(&req); err != nil {
		code := "[HANDLER] UpdatePassword - 3"
		log.Errorw(code, err)
		return respondError(ctx, fiber.StatusBadRequest, err.Error())
	}

	err := handler.userService.UpdatePassword(ctx.Context(), userID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, errors2.ErrorUserNotFound) {
			return respondError(ctx, fiber.StatusNotFound, "User not found")
		}

		if errors.Is(err, errors2.ErrorPasswordInvalid) {
			return respondError(ctx, fiber.StatusBadRequest, "Current password is invalid")
		}

		code := "[HANDLER] UpdatePassword - 4"
		log.Errorw(code, err)
		return respondError(ctx, fiber.StatusInternalServerError, "Failed to update password")
	}

	if handler.jwtAuth != nil {
		handler.jwtAuth.InvalidateUserTokens(userID)
		if newToken, expiredAt, tokenErr := handler.jwtAuth.GenerateToken(&entity.JwtData{UserID: userID}); tokenErr == nil {
			setAuthCookie(ctx, newToken, time.Unix(expiredAt, 0), isProductionEnv(handler.cfg))
		}
	}

	var successResp response.DefaultSuccessResponse
	successResp.Meta.Status = true
	successResp.Pagination = nil
	successResp.Meta.Message = "Password updated successfully"
	successResp.Data = nil
	return ctx.Status(fiber.StatusOK).JSON(successResp)
}

func (handler *userHandler) GetUserByID(ctx *fiber.Ctx) error {
	userID, ok := getAuthenticatedUserID(ctx)
	if !ok {
		code := "[HANDLER] GetUserByID - 1"
		log.Errorw(code, "reason", "unauthorized")
		return respondError(ctx, fiber.StatusUnauthorized, "Unauthorized")
	}

	user, err := handler.userService.GetUserByID(ctx.Context(), userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, errors2.ErrorUserNotFound) {
			return respondError(ctx, fiber.StatusNotFound, "User not found")
		}

		code := "[HANDLER] GetUserByID - 2"
		log.Errorw(code, err)
		return respondError(ctx, fiber.StatusInternalServerError, "Failed to get user")
	}

	var successResp response.DefaultSuccessResponse
	successResp.Meta.Status = true
	successResp.Pagination = nil
	successResp.Meta.Message = "User fetched successfully"
	successResp.Data = response.UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}
	return ctx.Status(fiber.StatusOK).JSON(successResp)
}
