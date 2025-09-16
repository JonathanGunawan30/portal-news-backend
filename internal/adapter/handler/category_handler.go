package handler

import (
	"errors"
	"strings"

	"github.com/JonathanGunawan30/portal-news-backend.git/internal/adapter/handler/request"
	"github.com/JonathanGunawan30/portal-news-backend.git/internal/adapter/handler/response"
	"github.com/JonathanGunawan30/portal-news-backend.git/internal/core/domain/entity"
	categoryError "github.com/JonathanGunawan30/portal-news-backend.git/internal/core/domain/errors"
	"github.com/JonathanGunawan30/portal-news-backend.git/internal/core/service"
	"github.com/JonathanGunawan30/portal-news-backend.git/lib/conv"
	validation "github.com/JonathanGunawan30/portal-news-backend.git/validator"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

type CategoryHandler interface {
	GetCategories(ctx *fiber.Ctx) error
	GetCategoryByID(ctx *fiber.Ctx) error
	CreateCategory(ctx *fiber.Ctx) error
	EditCategoryByID(ctx *fiber.Ctx) error
	DeleteCategoryByID(ctx *fiber.Ctx) error

	GetCategoryFE(ctx *fiber.Ctx) error
}

type categoryHandler struct {
	categoryService service.CategoryService
}

func NewCategoryHandler(categoryService service.CategoryService) CategoryHandler {
	return &categoryHandler{
		categoryService: categoryService,
	}
}

func (handler *categoryHandler) GetCategories(ctx *fiber.Ctx) error {
	claims, ok := ctx.Locals("user").(*entity.JwtData)
	userID := claims.UserID
	if userID == 0 || !ok {
		code := "[HANDLER] GetCategories - 1"
		log.Errorw(code, "reason", "unauthorized")
		errResp.Meta.Status = false
		errResp.Meta.Message = "Unauthorized"
		return ctx.Status(fiber.StatusUnauthorized).JSON(errResp)
	}

	results, err := handler.categoryService.GetCategories(ctx.Context())
	if err != nil {
		code := "[HANDLER] GetCategories - 2"
		log.Errorw(code, err)
		errResp.Meta.Status = false
		errResp.Meta.Message = "Failed to fetch categories"
		return ctx.Status(fiber.StatusInternalServerError).JSON(errResp)
	}

	var responses = make([]response.SuccessCategoryResponse, 0)
	for _, result := range results {
		responses = append(responses, response.SuccessCategoryResponse{
			ID:            result.ID,
			Title:         result.Title,
			Slug:          result.Slug,
			CreatedByName: result.User.Name,
		})
	}

	var successResp response.DefaultSuccessResponse
	successResp.Meta.Status = true
	successResp.Pagination = nil
	successResp.Meta.Message = "Categories fetched successfully"
	successResp.Data = responses
	return ctx.JSON(successResp)
}

func (handler *categoryHandler) GetCategoryByID(ctx *fiber.Ctx) error {
	claims, ok := ctx.Locals("user").(*entity.JwtData)
	userID := claims.UserID
	if userID == 0 || !ok {
		code := "[HANDLER] GetCategoryByID - 1"
		log.Errorw(code, "reason", "unauthorized")
		errResp.Meta.Status = false
		errResp.Meta.Message = "Unauthorized"
		return ctx.Status(fiber.StatusUnauthorized).JSON(errResp)
	}

	idStr := ctx.Params("id")
	id, err := conv.StringToInt64(idStr)
	if err != nil {
		code := "[HANDLER] GetCategoryByID - 2"
		log.Errorw(code, err)
		errResp.Meta.Status = false
		errResp.Meta.Message = "Invalid ID format"
		return ctx.Status(fiber.StatusBadRequest).JSON(errResp)
	}

	result, err := handler.categoryService.GetCategoryByID(ctx.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			code := "[HANDLER] GetCategoryByID - 2"
			log.Errorw(code, err)
			errResp.Meta.Status = false
			errResp.Meta.Message = "Category not found"
			return ctx.Status(fiber.StatusNotFound).JSON(errResp)
		}
		code := "[HANDLER] GetCategoryByID - 3"
		log.Errorw(code, err)
		errResp.Meta.Status = false
		errResp.Meta.Message = "Failed to get category"
		return ctx.Status(fiber.StatusInternalServerError).JSON(errResp)
	}

	categoryResponse := response.SuccessCategoryResponse{
		ID:            result.ID,
		Title:         result.Title,
		Slug:          result.Slug,
		CreatedByName: result.User.Name,
	}

	var successResp response.DefaultSuccessResponse
	successResp.Meta.Status = true
	successResp.Pagination = nil
	successResp.Meta.Message = "Category fetched successfully"
	successResp.Data = categoryResponse

	return ctx.JSON(successResp)

}

func (handler *categoryHandler) CreateCategory(ctx *fiber.Ctx) error {
	var req request.CategoryRequest
	claims, ok := ctx.Locals("user").(*entity.JwtData)
	userID := claims.UserID
	if userID == 0 || !ok {
		code := "[HANDLER] CreateCategory - 1"
		log.Errorw(code, "reason", "unauthorized")
		errResp.Meta.Status = false
		errResp.Meta.Message = "Unauthorized"
		return ctx.Status(fiber.StatusUnauthorized).JSON(errResp)
	}

	if err := ctx.BodyParser(&req); err != nil {
		code := "[HANDLER] CreateCategory - 2"
		log.Errorw(code, err)
		errResp.Meta.Status = false
		errResp.Meta.Message = "Failed to parse request body"
		return ctx.Status(fiber.StatusBadRequest).JSON(errResp)
	}

	if err := validation.ValidateStruct(&req); err != nil {
		code := "[HANDLER] CreateCategory - 3"
		log.Errorw(code, err)
		errResp.Meta.Status = false
		errResp.Meta.Message = err.Error()
		return ctx.Status(fiber.StatusBadRequest).JSON(errResp)
	}

	categoryRequest := entity.CategoryEntity{
		Title: req.Title,
		User: entity.UserEntity{
			ID: int64(userID),
		},
	}

	err := handler.categoryService.CreateCategory(ctx.Context(), categoryRequest)
	if err != nil {
		if errors.Is(err, categoryError.ErrorCategoryTitleAlreadyExist) {
			code := "[HANDLER] CreateCategory - 4"
			log.Errorw(code, err)
			errResp.Meta.Status = false
			errResp.Meta.Message = "Category title already exists"
			return ctx.Status(fiber.StatusConflict).JSON(errResp)
		}

		code := "[HANDLER] CreateCategory - 5"
		log.Errorw(code, err)
		errResp.Meta.Status = false
		errResp.Meta.Message = err.Error()
		return ctx.Status(fiber.StatusInternalServerError).JSON(errResp)
	}

	var successResp response.DefaultSuccessResponse
	successResp.Meta.Status = true
	successResp.Pagination = nil
	successResp.Meta.Message = "Category created successfully"
	successResp.Data = req
	return ctx.Status(fiber.StatusCreated).JSON(successResp)

}

func (handler *categoryHandler) EditCategoryByID(ctx *fiber.Ctx) error {
	var req request.CategoryRequest
	claims, ok := ctx.Locals("user").(*entity.JwtData)
	userID := claims.UserID
	if userID == 0 || !ok {
		code := "[HANDLER] EditCategoryByID - 1"
		log.Errorw(code, "reason", "unauthorized")
		errResp.Meta.Status = false
		errResp.Meta.Message = "Unauthorized"
		return ctx.Status(fiber.StatusUnauthorized).JSON(errResp)
	}

	idStr := ctx.Params("id")
	id, err := conv.StringToInt64(idStr)
	if err != nil {
		code := "[HANDLER] EditCategoryByID - 2"
		log.Errorw(code, err)
		errResp.Meta.Status = false
		errResp.Meta.Message = "Invalid ID format"
		return ctx.Status(fiber.StatusBadRequest).JSON(errResp)
	}

	if err = ctx.BodyParser(&req); err != nil {
		code := "[HANDLER] EditCategoryByID - 3"
		log.Errorw(code, err)
		errResp.Meta.Status = false
		errResp.Meta.Message = "Failed to parse request body"
		return ctx.Status(fiber.StatusBadRequest).JSON(errResp)
	}

	if err = validation.ValidateStruct(&req); err != nil {
		code := "[HANDLER] EditCategoryByID - 4"
		log.Errorw(code, err)
		errResp.Meta.Status = false
		errResp.Meta.Message = err.Error()
		return ctx.Status(fiber.StatusBadRequest).JSON(errResp)
	}

	categoryRequest := entity.CategoryEntity{
		ID:    id,
		Title: req.Title,
	}

	err = handler.categoryService.EditCategoryByID(ctx.Context(), categoryRequest)
	if err != nil {
		if errors.Is(err, categoryError.ErrorCategoryTitleAlreadyExist) {
			code := "[HANDLER] EditCategoryByID - 5"
			log.Errorw(code, err)
			errResp.Meta.Status = false
			errResp.Meta.Message = "Category title already exists"
			return ctx.Status(fiber.StatusConflict).JSON(errResp)
		}

		if errors.Is(err, categoryError.ErrorCategoryNotFoundToUpdate) {
			code := "[HANDLER] EditCategoryByID - 6"
			log.Errorw(code, err)
			errResp.Meta.Status = false
			errResp.Meta.Message = "Category not found"
			return ctx.Status(fiber.StatusNotFound).JSON(errResp)
		}

		code := "[HANDLER] EditCategoryByID - 7"
		log.Errorw(code, err)
		errResp.Meta.Status = false
		errResp.Meta.Message = err.Error()
		return ctx.Status(fiber.StatusInternalServerError).JSON(errResp)
	}

	var successResp response.DefaultSuccessResponse
	successResp.Meta.Status = true
	successResp.Pagination = nil
	successResp.Meta.Message = "Category updated successfully"
	successResp.Data = req
	return ctx.Status(fiber.StatusOK).JSON(successResp)
}

func (handler *categoryHandler) DeleteCategoryByID(ctx *fiber.Ctx) error {
	claims, ok := ctx.Locals("user").(*entity.JwtData)
	userID := claims.UserID
	if userID == 0 || !ok {
		code := "[HANDLER] DeleteCategoryByID - 1"
		log.Errorw(code, "reason", "unauthorized")
		errResp.Meta.Status = false
		errResp.Meta.Message = "Unauthorized"
		return ctx.Status(fiber.StatusUnauthorized).JSON(errResp)
	}

	idStr := ctx.Params("id")
	id, err := conv.StringToInt64(idStr)
	if err != nil {
		code := "[HANDLER] DeleteCategoryByID - 2"
		log.Errorw(code, err)
		errResp.Meta.Status = false
		errResp.Meta.Message = "Invalid ID format"
		return ctx.Status(fiber.StatusBadRequest).JSON(errResp)
	}

	err = handler.categoryService.DeleteCategoryByID(ctx.Context(), id)

	if err != nil {
		if errors.Is(err, categoryError.ErrorCategoryNotFoundToDelete) {
			code := "[HANDLER] DeleteCategoryByID - 3"
			log.Errorw(code, err)
			errResp.Meta.Status = false
			errResp.Meta.Message = "Category not found"
			return ctx.Status(fiber.StatusNotFound).JSON(errResp)
		}

		if errors.Is(err, categoryError.ErrorCategoryHasAssociatedContent) {
			code := "[HANDLER] DeleteCategoryByID - 4"
			log.Errorw(code, err)
			errResp.Meta.Status = false
			errResp.Meta.Message = "Cannot delete, category has associated content"
			return ctx.Status(fiber.StatusConflict).JSON(errResp)
		}

		code := "[HANDLER] DeleteCategoryByID - 5"
		log.Errorw(code, err)
		errResp.Meta.Status = false
		errResp.Meta.Message = err.Error()
		return ctx.Status(fiber.StatusInternalServerError).JSON(errResp)
	}

	var successResp response.DefaultSuccessResponse
	successResp.Meta.Status = true
	successResp.Pagination = nil
	successResp.Data = nil
	successResp.Meta.Message = "Category deleted successfully"
	return ctx.Status(fiber.StatusOK).JSON(successResp)

}

func (handler *categoryHandler) GetCategoryFE(ctx *fiber.Ctx) error {
	results, err := handler.categoryService.GetCategories(ctx.Context())
	if err != nil {
		code := "[HANDLER] GetCategoryFE - 1"
		log.Errorw(code, err)
		errResp.Meta.Status = false
		errResp.Meta.Message = "Failed to fetch categories"
		return ctx.Status(fiber.StatusInternalServerError).JSON(errResp)
	}

	categoryResponses := make([]response.SuccessCategoryResponse, 0)
	for _, result := range results {
		categoryResponses = append(categoryResponses, response.SuccessCategoryResponse{
			ID:            result.ID,
			Title:         result.Title,
			Slug:          result.Slug,
			CreatedByName: result.User.Name,
		})
	}

	var successResp response.DefaultSuccessResponse
	successResp.Meta.Status = true
	successResp.Pagination = nil
	successResp.Meta.Message = "Categories fetched successfully"
	successResp.Data = categoryResponses
	return ctx.JSON(successResp)

}
