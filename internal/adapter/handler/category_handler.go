package handler

import (
	"errors"

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
	if _, ok := getAuthenticatedUserID(ctx); !ok {
		code := "[HANDLER] GetCategories - 1"
		log.Errorw(code, "reason", "unauthorized")
		return respondError(ctx, fiber.StatusUnauthorized, "Unauthorized")
	}

	results, err := handler.categoryService.GetCategories(ctx.Context())
	if err != nil {
		code := "[HANDLER] GetCategories - 2"
		log.Errorw(code, err)
		return respondError(ctx, fiber.StatusInternalServerError, "Failed to fetch categories")
	}

	var responses = make([]response.SuccessCategoryResponse, 0, len(results))
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
	if _, ok := getAuthenticatedUserID(ctx); !ok {
		code := "[HANDLER] GetCategoryByID - 1"
		log.Errorw(code, "reason", "unauthorized")
		return respondError(ctx, fiber.StatusUnauthorized, "Unauthorized")
	}

	idStr := ctx.Params("id")
	id, err := conv.StringToInt64(idStr)
	if err != nil {
		code := "[HANDLER] GetCategoryByID - 2"
		log.Errorw(code, err)
		return respondError(ctx, fiber.StatusBadRequest, "Invalid ID format")
	}

	result, err := handler.categoryService.GetCategoryByID(ctx.Context(), id)
	if err != nil {
		if errors.Is(err, categoryError.ErrorCategoryNotFound) {
			code := "[HANDLER] GetCategoryByID - 3"
			log.Errorw(code, err)
			return respondError(ctx, fiber.StatusNotFound, "Category not found")
		}
		code := "[HANDLER] GetCategoryByID - 4"
		log.Errorw(code, err)
		return respondError(ctx, fiber.StatusInternalServerError, "Failed to get category")
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
	userID, ok := getAuthenticatedUserID(ctx)
	if !ok {
		code := "[HANDLER] CreateCategory - 1"
		log.Errorw(code, "reason", "unauthorized")
		return respondError(ctx, fiber.StatusUnauthorized, "Unauthorized")
	}

	if err := ctx.BodyParser(&req); err != nil {
		code := "[HANDLER] CreateCategory - 2"
		log.Errorw(code, err)
		return respondError(ctx, fiber.StatusBadRequest, "Failed to parse request body")
	}

	if err := validation.ValidateStruct(&req); err != nil {
		code := "[HANDLER] CreateCategory - 3"
		log.Errorw(code, err)
		return respondError(ctx, fiber.StatusBadRequest, err.Error())
	}

	categoryRequest := entity.CategoryEntity{
		Title: req.Title,
		User: entity.UserEntity{
			ID: userID,
		},
	}

	err := handler.categoryService.CreateCategory(ctx.Context(), categoryRequest)
	if err != nil {
		if errors.Is(err, categoryError.ErrorCategoryTitleAlreadyExist) {
			code := "[HANDLER] CreateCategory - 4"
			log.Errorw(code, err)
			return respondError(ctx, fiber.StatusConflict, "Category title already exists")
		}

		code := "[HANDLER] CreateCategory - 5"
		log.Errorw(code, err)
		return respondError(ctx, fiber.StatusInternalServerError, "Failed to create category")
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
	if _, ok := getAuthenticatedUserID(ctx); !ok {
		code := "[HANDLER] EditCategoryByID - 1"
		log.Errorw(code, "reason", "unauthorized")
		return respondError(ctx, fiber.StatusUnauthorized, "Unauthorized")
	}

	idStr := ctx.Params("id")
	id, err := conv.StringToInt64(idStr)
	if err != nil {
		code := "[HANDLER] EditCategoryByID - 2"
		log.Errorw(code, err)
		return respondError(ctx, fiber.StatusBadRequest, "Invalid ID format")
	}

	if err = ctx.BodyParser(&req); err != nil {
		code := "[HANDLER] EditCategoryByID - 3"
		log.Errorw(code, err)
		return respondError(ctx, fiber.StatusBadRequest, "Failed to parse request body")
	}

	if err = validation.ValidateStruct(&req); err != nil {
		code := "[HANDLER] EditCategoryByID - 4"
		log.Errorw(code, err)
		return respondError(ctx, fiber.StatusBadRequest, err.Error())
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
			return respondError(ctx, fiber.StatusConflict, "Category title already exists")
		}

		if errors.Is(err, categoryError.ErrorCategoryNotFoundToUpdate) {
			code := "[HANDLER] EditCategoryByID - 6"
			log.Errorw(code, err)
			return respondError(ctx, fiber.StatusNotFound, "Category not found")
		}

		code := "[HANDLER] EditCategoryByID - 7"
		log.Errorw(code, err)
		return respondError(ctx, fiber.StatusInternalServerError, "Failed to update category")
	}

	var successResp response.DefaultSuccessResponse
	successResp.Meta.Status = true
	successResp.Pagination = nil
	successResp.Meta.Message = "Category updated successfully"
	successResp.Data = req
	return ctx.Status(fiber.StatusOK).JSON(successResp)
}

func (handler *categoryHandler) DeleteCategoryByID(ctx *fiber.Ctx) error {
	if _, ok := getAuthenticatedUserID(ctx); !ok {
		code := "[HANDLER] DeleteCategoryByID - 1"
		log.Errorw(code, "reason", "unauthorized")
		return respondError(ctx, fiber.StatusUnauthorized, "Unauthorized")
	}

	idStr := ctx.Params("id")
	id, err := conv.StringToInt64(idStr)
	if err != nil {
		code := "[HANDLER] DeleteCategoryByID - 2"
		log.Errorw(code, err)
		return respondError(ctx, fiber.StatusBadRequest, "Invalid ID format")
	}

	err = handler.categoryService.DeleteCategoryByID(ctx.Context(), id)
	if err != nil {
		if errors.Is(err, categoryError.ErrorCategoryNotFoundToDelete) {
			code := "[HANDLER] DeleteCategoryByID - 3"
			log.Errorw(code, err)
			return respondError(ctx, fiber.StatusNotFound, "Category not found")
		}

		if errors.Is(err, categoryError.ErrorCategoryHasAssociatedContent) {
			code := "[HANDLER] DeleteCategoryByID - 4"
			log.Errorw(code, err)
			return respondError(ctx, fiber.StatusConflict, "Cannot delete, category has associated content")
		}

		code := "[HANDLER] DeleteCategoryByID - 5"
		log.Errorw(code, err)
		return respondError(ctx, fiber.StatusInternalServerError, "Failed to delete category")
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
		return respondError(ctx, fiber.StatusInternalServerError, "Failed to fetch categories")
	}

	categoryResponses := make([]response.SuccessCategoryResponse, 0, len(results))
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
