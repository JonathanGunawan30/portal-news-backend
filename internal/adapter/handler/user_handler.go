package handler

import (
	"errors"

	"github.com/JonathanGunawan30/portal-news-backend.git/internal/adapter/handler/request"
	"github.com/JonathanGunawan30/portal-news-backend.git/internal/adapter/handler/response"
	"github.com/JonathanGunawan30/portal-news-backend.git/internal/core/domain/entity"
	errors2 "github.com/JonathanGunawan30/portal-news-backend.git/internal/core/domain/errors"
	"github.com/JonathanGunawan30/portal-news-backend.git/internal/core/service"
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
}

func NewUserHandler(userService service.UserService) UserHandler {
	return &userHandler{
		userService: userService,
	}
}

func (handler *userHandler) UpdatePassword(ctx *fiber.Ctx) error {
	claims, ok := ctx.Locals("user").(*entity.JwtData)
	if claims.UserID == 0 || !ok {
		code := "[HANDLER] GetUserByID - 1"
		log.Errorw(code, "reason", "unauthorized")
		errResp.Meta.Status = false
		errResp.Meta.Message = "Unauthorized"
		return ctx.Status(fiber.StatusUnauthorized).JSON(errResp)
	}

	userID := claims.UserID
	var req request.UpdatePasswordRequest
	if err := ctx.BodyParser(&req); err != nil {
		code := "[HANDLER] UpdatePassword - 2"
		log.Errorw(code, err)
		errResp.Meta.Status = false
		errResp.Meta.Message = "Failed to parse request body"
		return ctx.Status(fiber.StatusBadRequest).JSON(errResp)
	}
	if err := validation.ValidateStruct(&req); err != nil {
		code := "[HANDLER] UpdatePassword - 3"
		log.Errorw(code, err)
		errResp.Meta.Status = false
		errResp.Meta.Message = err.Error()
		return ctx.Status(fiber.StatusBadRequest).JSON(errResp)
	}

	err := handler.userService.UpdatePassword(ctx.Context(), int64(userID), req.CurrentPassword, req.NewPassword)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			errResp.Meta.Status = false
			errResp.Meta.Message = "User not found"
			return ctx.Status(fiber.StatusNotFound).JSON(errResp)
		}

		if errors.Is(err, errors2.ErrorPasswordInvalid) {
			errResp.Meta.Status = false
			errResp.Meta.Message = "Current password is invalid"
			return ctx.Status(fiber.StatusBadRequest).JSON(errResp)
		}

		code := "[HANDLER] UpdatePassword - 4"
		log.Errorw(code, err)
		errResp.Meta.Status = false
		errResp.Meta.Message = "Failed to update password"
		return ctx.Status(fiber.StatusInternalServerError).JSON(errResp)
	}

	var successResp response.DefaultSuccessResponse
	successResp.Meta.Status = true
	successResp.Pagination = nil
	successResp.Meta.Message = "Password updated successfully"
	successResp.Data = nil
	return ctx.Status(fiber.StatusOK).JSON(successResp)

}

func (handler *userHandler) GetUserByID(ctx *fiber.Ctx) error {
	claims, ok := ctx.Locals("user").(*entity.JwtData)
	if claims.UserID == 0 || !ok {
		code := "[HANDLER] GetUserByID - 1"
		log.Errorw(code, "reason", "unauthorized")
		errResp.Meta.Status = false
		errResp.Meta.Message = "Unauthorized"
		return ctx.Status(fiber.StatusUnauthorized).JSON(errResp)
	}

	userID := claims.UserID
	user, err := handler.userService.GetUserByID(ctx.Context(), int64(userID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			errResp.Meta.Status = false
			errResp.Meta.Message = "User not found"
			return ctx.Status(fiber.StatusNotFound).JSON(errResp)
		}

		code := "[HANDLER] GetUserByID - 2"
		log.Errorw(code, err)
		errResp.Meta.Status = false
		errResp.Meta.Message = "Failed to get user"
		return ctx.Status(fiber.StatusInternalServerError).JSON(errResp)
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
