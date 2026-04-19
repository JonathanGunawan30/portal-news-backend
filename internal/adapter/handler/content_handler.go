package handler

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/JonathanGunawan30/portal-news-backend.git/internal/adapter/handler/request"
	"github.com/JonathanGunawan30/portal-news-backend.git/internal/adapter/handler/response"
	"github.com/JonathanGunawan30/portal-news-backend.git/internal/core/domain/entity"
	contentError "github.com/JonathanGunawan30/portal-news-backend.git/internal/core/domain/errors"
	"github.com/JonathanGunawan30/portal-news-backend.git/internal/core/service"
	"github.com/JonathanGunawan30/portal-news-backend.git/lib/conv"
	validation "github.com/JonathanGunawan30/portal-news-backend.git/validator"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

type ContentHandler interface {
	GetContents(ctx *fiber.Ctx) error
	GetContentByID(ctx *fiber.Ctx) error
	CreateContent(ctx *fiber.Ctx) error
	EditContentByID(ctx *fiber.Ctx) error
	DeleteContentByID(ctx *fiber.Ctx) error
	UploadImageR2(ctx *fiber.Ctx) error

	GetContentWithQuery(ctx *fiber.Ctx) error
	GetContentDetail(ctx *fiber.Ctx) error
}

type contentHandler struct {
	contentService service.ContentService
}

func NewContentHandler(contentService service.ContentService) ContentHandler {
	return &contentHandler{
		contentService: contentService,
	}
}

func (handler *contentHandler) GetContents(ctx *fiber.Ctx) error {
	claims, ok := ctx.Locals("user").(*entity.JwtData)
	if claims.UserID == 0 || !ok {
		code := "[HANDLER] GetContents - 1"
		log.Errorw(code, "reason", "unauthorized")
		errResp.Meta.Status = false
		errResp.Meta.Message = "Unauthorized"
		return ctx.Status(fiber.StatusUnauthorized).JSON(errResp)
	}

	var err error
	page := 1
	if ctx.Query("page") != "" {
		page, err = conv.StringToInt(ctx.Query("page"))
		if err != nil {
			code := "[HANDLER] GetContents - 2"
			log.Errorw(code, err)
			errResp.Meta.Status = false
			errResp.Meta.Message = "Invalid page format"
			return ctx.Status(fiber.StatusBadRequest).JSON(errResp)
		}
	}

	limit := 10
	if ctx.Query("limit") != "" {
		limit, err = conv.StringToInt(ctx.Query("limit"))
		if err != nil {
			code := "[HANDLER] GetContents - 3"
			log.Errorw(code, err)
			errResp.Meta.Status = false
			errResp.Meta.Message = "Invalid limit format"
			return ctx.Status(fiber.StatusBadRequest).JSON(errResp)
		}
	}

	order := "created_at"
	if ctx.Query("order") != "" {
		order = ctx.Query("orderBy")
	}

	orderType := "DESC"
	if ctx.Query("orderType") != "" {
		orderType = ctx.Query("orderType")
	}

	search := ""
	if ctx.Query("search") != "" {
		search = ctx.Query("search")
	}

	categoryID := 0
	if ctx.Query("category_id") != "" {
		categoryID, err = conv.StringToInt(ctx.Query("category_id"))
		if err != nil {
			code := "[HANDLER] GetContents - 3"
			log.Errorw(code, err)
			errResp.Meta.Status = false
			errResp.Meta.Message = "Invalid category_id format"
			return ctx.Status(fiber.StatusBadRequest).JSON(errResp)
		}
	}

	reqEntity := entity.QueryString{
		Limit:      limit,
		Page:       page,
		OrderBy:    order,
		OrderType:  orderType,
		Search:     search,
		CategoryID: int64(categoryID),
	}

	results, totalData, totalPages, err := handler.contentService.GetContents(ctx.Context(), reqEntity)

	if err != nil {
		code := "[HANDLER] GetContents - 4"
		log.Errorw(code, err)
		errResp.Meta.Status = false
		errResp.Meta.Message = err.Error()
		return ctx.Status(fiber.StatusInternalServerError).JSON(errResp)
	}

	var responses = make([]response.ContentResponse, 0)

	for _, result := range results {
		contentResponse := response.ContentResponse{
			ID:           result.ID,
			Title:        result.Title,
			Excerpt:      result.Excerpt,
			Description:  result.Description,
			Image:        result.Image,
			Tags:         result.Tags,
			Status:       result.Status,
			CategoryID:   result.CategoryID,
			CreatedByID:  result.CreatedByID,
			CreatedAt:    result.CreatedAt.Format(time.RFC3339),
			UpdatedAt:    result.UpdatedAt.Format(time.RFC3339),
			CategoryName: result.Category.Title,
			Author:       result.User.Name,
		}

		responses = append(responses, contentResponse)
	}

	var successResp response.DefaultSuccessResponse
	successResp.Meta.Status = true
	successResp.Pagination = &response.PaginationResponse{
		TotalRecord: totalData,
		TotalPage:   int64(totalPages),
		Page:        int64(page),
		PerPage:     int64(limit),
	}
	successResp.Meta.Message = "Contents fetched successfully"
	successResp.Data = responses
	return ctx.JSON(successResp)
}

func (handler *contentHandler) GetContentByID(ctx *fiber.Ctx) error {
	claims, ok := ctx.Locals("user").(*entity.JwtData)
	if claims.UserID == 0 || !ok {
		code := "[HANDLER] GetContentByID - 1"
		log.Errorw(code, "reason", "unauthorized")
		errResp.Meta.Status = false
		errResp.Meta.Message = "Unauthorized"
		return ctx.Status(fiber.StatusUnauthorized).JSON(errResp)
	}

	idParams := ctx.Params("id")
	id, err := conv.StringToInt64(idParams)
	if err != nil {
		code := "[HANDLER] GetContentByID - 2"
		log.Errorw(code, err)
		errResp.Meta.Status = false
		errResp.Meta.Message = "Invalid ID format"
		return ctx.Status(fiber.StatusBadRequest).JSON(errResp)
	}

	result, err := handler.contentService.GetContentByID(ctx.Context(), id)
	if err != nil {
		if errors.Is(err, contentError.ErrorContentNotFound) {
			code := "[HANDLER] GetContentByID - 3"
			log.Errorw(code, err)
			errResp.Meta.Status = false
			errResp.Meta.Message = "Content not found"
			return ctx.Status(fiber.StatusNotFound).JSON(errResp)
		}

		code := "[HANDLER] GetContentByID - 4"
		log.Errorw(code, err)
		errResp.Meta.Status = false
		errResp.Meta.Message = "Failed to get content"
		return ctx.Status(fiber.StatusInternalServerError).JSON(errResp)
	}

	contentResponse := response.ContentResponse{
		ID:           result.ID,
		Title:        result.Title,
		Excerpt:      result.Excerpt,
		Description:  result.Description,
		Image:        result.Image,
		Tags:         result.Tags,
		Status:       result.Status,
		CategoryID:   result.CategoryID,
		CreatedByID:  result.CreatedByID,
		CreatedAt:    result.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    result.UpdatedAt.Format(time.RFC3339),
		CategoryName: result.Category.Title,
		Author:       result.User.Name,
	}

	var successResp response.DefaultSuccessResponse
	successResp.Meta.Status = true
	successResp.Pagination = nil
	successResp.Meta.Message = "Content fetched successfully"
	successResp.Data = contentResponse
	return ctx.JSON(successResp)

}

func (handler *contentHandler) CreateContent(ctx *fiber.Ctx) error {
	claims, ok := ctx.Locals("user").(*entity.JwtData)
	if claims.UserID == 0 || !ok {
		code := "[HANDLER] CreateContent - 1"
		log.Errorw(code, "reason", "unauthorized")
		errResp.Meta.Status = false
		errResp.Meta.Message = "Unauthorized"
		return ctx.Status(fiber.StatusUnauthorized).JSON(errResp)
	}

	userID := claims.UserID
	var contentRequest request.ContentRequest
	if err := ctx.BodyParser(&contentRequest); err != nil {
		code := "[HANDLER] CreateContent - 2"
		log.Errorw(code, err)
		errResp.Meta.Status = false
		errResp.Meta.Message = "Failed to parse request body"
		return ctx.Status(fiber.StatusBadRequest).JSON(errResp)
	}

	if err := validation.ValidateStruct(&contentRequest); err != nil {
		code := "[HANDLER] CreateContent - 3"
		log.Errorw(code, err)
		errResp.Meta.Status = false
		errResp.Meta.Message = err.Error()
		return ctx.Status(fiber.StatusBadRequest).JSON(errResp)
	}

	tags := strings.Split(contentRequest.Tags, ",")
	reqEntity := entity.ContentEntity{
		Title:       contentRequest.Title,
		Excerpt:     contentRequest.Excerpt,
		Description: contentRequest.Description,
		Image:       contentRequest.Image,
		Tags:        tags,
		Status:      contentRequest.Status,
		CategoryID:  contentRequest.CategoryID,
		CreatedByID: int64(userID),
	}

	err := handler.contentService.CreateContent(ctx.Context(), reqEntity)
	if err != nil {
		code := "[HANDLER] CreateContent - 4"
		log.Errorw(code, err)
		errResp.Meta.Status = false
		errResp.Meta.Message = err.Error()
		return ctx.Status(fiber.StatusBadRequest).JSON(errResp)
	}
	var successResp response.DefaultSuccessResponse
	successResp.Meta.Status = true
	successResp.Pagination = nil
	successResp.Meta.Message = "Content created successfully"
	return ctx.Status(fiber.StatusCreated).JSON(successResp)

}

func (handler *contentHandler) EditContentByID(ctx *fiber.Ctx) error {
	claims, ok := ctx.Locals("user").(*entity.JwtData)
	if claims.UserID == 0 || !ok {
		code := "[HANDLER] UpdateContent - 1"
		log.Errorw(code, "reason", "unauthorized")
		errResp.Meta.Status = false
		errResp.Meta.Message = "Unauthorized"
		return ctx.Status(fiber.StatusUnauthorized).JSON(errResp)
	}

	userID := claims.UserID
	var contentRequest request.ContentRequest
	if err := ctx.BodyParser(&contentRequest); err != nil {
		code := "[HANDLER] UpdateContent - 2"
		log.Errorw(code, err)
		errResp.Meta.Status = false
		errResp.Meta.Message = "Failed to parse request body"
		return ctx.Status(fiber.StatusBadRequest).JSON(errResp)
	}

	if err := validation.ValidateStruct(&contentRequest); err != nil {
		code := "[HANDLER] UpdateContent - 3"
		log.Errorw(code, err)
		errResp.Meta.Status = false
		errResp.Meta.Message = err.Error()
		return ctx.Status(fiber.StatusBadRequest).JSON(errResp)
	}

	idParams := ctx.Params("id")
	id, err := conv.StringToInt64(idParams)
	if err != nil {
		code := "[HANDLER] UpdateContent - 4"
		log.Errorw(code, err)
		errResp.Meta.Status = false
		errResp.Meta.Message = "Invalid ID format"
		return ctx.Status(fiber.StatusBadRequest).JSON(errResp)
	}

	tags := strings.Split(contentRequest.Tags, ",")
	reqEntity := entity.ContentEntity{
		ID:          id,
		Title:       contentRequest.Title,
		Excerpt:     contentRequest.Excerpt,
		Description: contentRequest.Description,
		Image:       contentRequest.Image,
		Tags:        tags,
		Status:      contentRequest.Status,
		CategoryID:  contentRequest.CategoryID,
		CreatedByID: int64(userID),
	}

	err = handler.contentService.UpdateContent(ctx.Context(), reqEntity)
	if err != nil {
		if errors.Is(err, contentError.ErrorContentNotFound) {
			code := "[HANDLER] UpdateContent - 5"
			log.Errorw(code, err)
			errResp.Meta.Status = false
			errResp.Meta.Message = "Content not found"
			return ctx.Status(fiber.StatusNotFound).JSON(errResp)
		}
		code := "[HANDLER] UpdateContent - 6"
		log.Errorw(code, err)
		errResp.Meta.Status = false
		errResp.Meta.Message = "Failed to update content"
		return ctx.Status(fiber.StatusInternalServerError).JSON(errResp)
	}

	var successResp response.DefaultSuccessResponse
	successResp.Meta.Status = true
	successResp.Pagination = nil
	successResp.Meta.Message = "Content updated successfully"
	return ctx.Status(fiber.StatusOK).JSON(successResp)
}

func (handler *contentHandler) DeleteContentByID(ctx *fiber.Ctx) error {
	claims, ok := ctx.Locals("user").(*entity.JwtData)
	if claims.UserID == 0 || !ok {
		code := "[HANDLER] DeleteContentByID - 1"
		log.Errorw(code, "reason", "unauthorized")
		errResp.Meta.Status = false
		errResp.Meta.Message = "Unauthorized"
		return ctx.Status(fiber.StatusUnauthorized).JSON(errResp)
	}

	idParams := ctx.Params("id")
	id, err := conv.StringToInt64(idParams)
	if err != nil {
		code := "[HANDLER] DeleteContentByID - 2"
		log.Errorw(code, err)
		errResp.Meta.Status = false
		errResp.Meta.Message = "Invalid ID format"
		return ctx.Status(fiber.StatusBadRequest).JSON(errResp)
	}

	err = handler.contentService.DeleteContentByID(ctx.Context(), id)
	if err != nil {
		if errors.Is(err, contentError.ErrorContentNotFound) {
			code := "[HANDLER] DeleteContentByID - 3"
			log.Errorw(code, err)
			errResp.Meta.Status = false
			errResp.Meta.Message = "Content not found"
			return ctx.Status(fiber.StatusNotFound).JSON(errResp)
		}
		code := "[HANDLER] DeleteContentByID - 4"
		log.Errorw(code, err)
		errResp.Meta.Status = false
		errResp.Meta.Message = "Failed to delete content"
		return ctx.Status(fiber.StatusInternalServerError).JSON(errResp)
	}

	var successResp response.DefaultSuccessResponse
	successResp.Meta.Status = true
	successResp.Pagination = nil
	successResp.Meta.Message = "Content deleted successfully"
	return ctx.Status(fiber.StatusOK).JSON(successResp)
}
func (handler *contentHandler) UploadImageR2(ctx *fiber.Ctx) error {
	claims, ok := ctx.Locals("user").(*entity.JwtData)
	if !ok || claims.UserID == 0 {
		code := "[HANDLER] UploadImageR2 - 1"
		log.Errorw(code, "reason", "unauthorized")
		errResp.Meta.Status = false
		errResp.Meta.Message = "Unauthorized"
		return ctx.Status(fiber.StatusUnauthorized).JSON(errResp)
	}

	file, err := ctx.FormFile("image")
	if err != nil {
		code := "[HANDLER] UploadImageR2 - 2"
		log.Errorw(code, err)
		errResp.Meta.Status = false
		errResp.Meta.Message = "Failed to parse request body"
		return ctx.Status(fiber.StatusBadRequest).JSON(errResp)
	}

	f, err := file.Open()
	if err != nil {
		code := "[HANDLER] UploadImageR2 - 3"
		log.Errorw(code, err)
		errResp.Meta.Status = false
		errResp.Meta.Message = "Failed to open uploaded file"
		return ctx.Status(fiber.StatusBadRequest).JSON(errResp)
	}
	defer f.Close()

	fileName := fmt.Sprintf("%d-%d-%s", int64(claims.UserID), time.Now().UnixNano(), file.Filename)

	imageURL, err := handler.contentService.UploadImageR2(ctx.Context(), entity.FileUploadEntity{
		Name: fileName,
	}, f)

	if err != nil {
		code := "[HANDLER] UploadImageR2 - 4"
		log.Errorw(code, err)
		errResp.Meta.Status = false
		errResp.Meta.Message = "Failed to upload image"
		return ctx.Status(fiber.StatusInternalServerError).JSON(errResp)
	}

	var successResp response.DefaultSuccessResponse
	successResp.Meta.Status = true
	successResp.Meta.Message = "Image uploaded successfully"
	successResp.Data = map[string]string{
		"url_image": imageURL,
	}
	return ctx.Status(fiber.StatusCreated).JSON(successResp)
}

func (handler *contentHandler) GetContentWithQuery(ctx *fiber.Ctx) error {
	var err error
	page := 1
	if ctx.Query("page") != "" {
		page, err = conv.StringToInt(ctx.Query("page"))
		if err != nil {
			code := "[HANDLER] GetContentWithQuery - 1"
			log.Errorw(code, err)
			errResp.Meta.Status = false
			errResp.Meta.Message = "Invalid page format"
			return ctx.Status(fiber.StatusBadRequest).JSON(errResp)
		}
	}

	limit := 6
	if ctx.Query("limit") != "" {
		limit, err = conv.StringToInt(ctx.Query("limit"))
		if err != nil {
			code := "[HANDLER] GetContentWithQuery - 2"
			log.Errorw(code, err)
			errResp.Meta.Status = false
			errResp.Meta.Message = "Invalid limit format"
			return ctx.Status(fiber.StatusBadRequest).JSON(errResp)
		}
	}

	order := "created_at"
	if ctx.Query("order") != "" {
		order = ctx.Query("orderBy")
	}

	orderType := "DESC"
	if ctx.Query("orderType") != "" {
		orderType = ctx.Query("orderType")
	}

	search := ""
	if ctx.Query("search") != "" {
		search = ctx.Query("search")
	}

	categoryID := 0
	if ctx.Query("category_id") != "" {
		categoryID, err = conv.StringToInt(ctx.Query("category_id"))
		if err != nil {
			code := "[HANDLER] GetContentWithQuery - 2"
			log.Errorw(code, err)
			errResp.Meta.Status = false
			errResp.Meta.Message = "Invalid category_id format"
			return ctx.Status(fiber.StatusBadRequest).JSON(errResp)
		}
	}

	reqEntity := entity.QueryString{
		Limit:      limit,
		Page:       page,
		OrderBy:    order,
		OrderType:  orderType,
		Search:     search,
		CategoryID: int64(categoryID),
		Status:     "PUBLISH",
	}

	results, totalData, totalPages, err := handler.contentService.GetContents(ctx.Context(), reqEntity)
	if err != nil {
		code := "[HANDLER] GetContentWithQuery - 3"
		log.Errorw(code, err)
		errResp.Meta.Status = false
		errResp.Meta.Message = err.Error()
		return ctx.Status(fiber.StatusInternalServerError).JSON(errResp)
	}

	var responses = make([]response.ContentResponse, 0)

	for _, result := range results {
		contentResponse := response.ContentResponse{
			ID:           result.ID,
			Title:        result.Title,
			Excerpt:      result.Excerpt,
			Description:  result.Description,
			Image:        result.Image,
			Tags:         result.Tags,
			Status:       result.Status,
			CategoryID:   result.CategoryID,
			CreatedByID:  result.CreatedByID,
			CreatedAt:    result.CreatedAt.Format(time.RFC3339),
			UpdatedAt:    result.UpdatedAt.Format(time.RFC3339),
			CategoryName: result.Category.Title,
			Author:       result.User.Name,
		}

		responses = append(responses, contentResponse)
	}

	var successResp response.DefaultSuccessResponse
	successResp.Meta.Status = true
	successResp.Pagination = &response.PaginationResponse{
		TotalRecord: totalData,
		TotalPage:   int64(totalPages),
		Page:        int64(page),
		PerPage:     int64(limit),
	}
	successResp.Meta.Message = "Contents fetched successfully"
	successResp.Data = responses
	return ctx.JSON(successResp)
}

func (handler *contentHandler) GetContentDetail(ctx *fiber.Ctx) error {
	idParams := ctx.Params("id")
	id, err := conv.StringToInt64(idParams)
	if err != nil {
		code := "[HANDLER] GetContentDetail - 1"
		log.Errorw(code, err)
		errResp.Meta.Status = false
		errResp.Meta.Message = "Invalid ID format"
		return ctx.Status(fiber.StatusBadRequest).JSON(errResp)
	}

	result, err := handler.contentService.GetContentByID(ctx.Context(), id)
	if err != nil {
		if errors.Is(err, contentError.ErrorContentNotFound) {
			code := "[HANDLER] GetContentDetail - 2"
			log.Errorw(code, err)
			errResp.Meta.Status = false
			errResp.Meta.Message = "Content not found"
			return ctx.Status(fiber.StatusNotFound).JSON(errResp)
		}

		code := "[HANDLER] GetContentDetail - 3"
		log.Errorw(code, err)
		errResp.Meta.Status = false
		errResp.Meta.Message = "Failed to get content"
		return ctx.Status(fiber.StatusInternalServerError).JSON(errResp)
	}

	if result.Status != "PUBLISH" {
		code := "[HANDLER] GetContentDetail - 4"
		log.Errorw(code, "reason", "content not published")
		errResp.Meta.Status = false
		errResp.Meta.Message = "Content not found"
		return ctx.Status(fiber.StatusNotFound).JSON(errResp)
	}

	contentResponse := response.ContentResponse{
		ID:           result.ID,
		Title:        result.Title,
		Excerpt:      result.Excerpt,
		Description:  result.Description,
		Image:        result.Image,
		Tags:         result.Tags,
		Status:       result.Status,
		CategoryID:   result.CategoryID,
		CreatedByID:  result.CreatedByID,
		CreatedAt:    result.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    result.UpdatedAt.Format(time.RFC3339),
		CategoryName: result.Category.Title,
		Author:       result.User.Name,
	}

	var successResp response.DefaultSuccessResponse
	successResp.Meta.Status = true
	successResp.Pagination = nil
	successResp.Meta.Message = "Content fetched successfully"
	successResp.Data = contentResponse
	return ctx.JSON(successResp)

}
