package handler

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
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
	"github.com/google/uuid"
)

const maxImageUploadSize = 5 * 1024 * 1024 // 5 MB

var allowedImageTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/gif":  ".gif",
}

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

func mapContentResponse(result entity.ContentEntity) response.ContentResponse {
	return response.ContentResponse{
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
}

func parseTags(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return []string{}
	}
	parts := strings.Split(raw, ",")
	tags := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			tags = append(tags, trimmed)
		}
	}
	return tags
}

func (handler *contentHandler) GetContents(ctx *fiber.Ctx) error {
	if _, ok := getAuthenticatedUserID(ctx); !ok {
		code := "[HANDLER] GetContents - 1"
		log.Errorw(code, "reason", "unauthorized")
		return respondError(ctx, fiber.StatusUnauthorized, "Unauthorized")
	}

	reqEntity, err := parseContentQuery(ctx, 10, false)
	if err != nil {
		code := "[HANDLER] GetContents - 2"
		log.Errorw(code, err)
		return respondError(ctx, fiber.StatusBadRequest, err.Error())
	}

	results, totalData, totalPages, err := handler.contentService.GetContents(ctx.Context(), reqEntity)
	if err != nil {
		code := "[HANDLER] GetContents - 3"
		log.Errorw(code, err)
		return respondError(ctx, fiber.StatusInternalServerError, "Failed to fetch contents")
	}

	responses := make([]response.ContentResponse, 0, len(results))
	for _, result := range results {
		responses = append(responses, mapContentResponse(result))
	}

	var successResp response.DefaultSuccessResponse
	successResp.Meta.Status = true
	successResp.Pagination = &response.PaginationResponse{
		TotalRecord: totalData,
		TotalPage:   int64(totalPages),
		Page:        int64(reqEntity.Page),
		PerPage:     int64(reqEntity.Limit),
	}
	successResp.Meta.Message = "Contents fetched successfully"
	successResp.Data = responses
	return ctx.JSON(successResp)
}

func (handler *contentHandler) GetContentByID(ctx *fiber.Ctx) error {
	if _, ok := getAuthenticatedUserID(ctx); !ok {
		code := "[HANDLER] GetContentByID - 1"
		log.Errorw(code, "reason", "unauthorized")
		return respondError(ctx, fiber.StatusUnauthorized, "Unauthorized")
	}

	idParams := ctx.Params("id")
	id, err := conv.StringToInt64(idParams)
	if err != nil {
		code := "[HANDLER] GetContentByID - 2"
		log.Errorw(code, err)
		return respondError(ctx, fiber.StatusBadRequest, "Invalid ID format")
	}

	result, err := handler.contentService.GetContentByID(ctx.Context(), id)
	if err != nil {
		if errors.Is(err, contentError.ErrorContentNotFound) {
			code := "[HANDLER] GetContentByID - 3"
			log.Errorw(code, err)
			return respondError(ctx, fiber.StatusNotFound, "Content not found")
		}

		code := "[HANDLER] GetContentByID - 4"
		log.Errorw(code, err)
		return respondError(ctx, fiber.StatusInternalServerError, "Failed to get content")
	}

	var successResp response.DefaultSuccessResponse
	successResp.Meta.Status = true
	successResp.Pagination = nil
	successResp.Meta.Message = "Content fetched successfully"
	successResp.Data = mapContentResponse(*result)
	return ctx.JSON(successResp)
}

func (handler *contentHandler) CreateContent(ctx *fiber.Ctx) error {
	userID, ok := getAuthenticatedUserID(ctx)
	if !ok {
		code := "[HANDLER] CreateContent - 1"
		log.Errorw(code, "reason", "unauthorized")
		return respondError(ctx, fiber.StatusUnauthorized, "Unauthorized")
	}

	var contentRequest request.ContentRequest
	if err := ctx.BodyParser(&contentRequest); err != nil {
		code := "[HANDLER] CreateContent - 2"
		log.Errorw(code, err)
		return respondError(ctx, fiber.StatusBadRequest, "Failed to parse request body")
	}

	if err := validation.ValidateStruct(&contentRequest); err != nil {
		code := "[HANDLER] CreateContent - 3"
		log.Errorw(code, err)
		return respondError(ctx, fiber.StatusBadRequest, err.Error())
	}

	status := contentRequest.Status
	if status == "" {
		status = "PUBLISH"
	}

	reqEntity := entity.ContentEntity{
		Title:       contentRequest.Title,
		Excerpt:     contentRequest.Excerpt,
		Description: contentRequest.Description,
		Image:       contentRequest.Image,
		Tags:        parseTags(contentRequest.Tags),
		Status:      status,
		CategoryID:  contentRequest.CategoryID,
		CreatedByID: userID,
	}

	err := handler.contentService.CreateContent(ctx.Context(), reqEntity)
	if err != nil {
		code := "[HANDLER] CreateContent - 4"
		log.Errorw(code, err)
		return respondError(ctx, fiber.StatusBadRequest, "Failed to create content")
	}
	var successResp response.DefaultSuccessResponse
	successResp.Meta.Status = true
	successResp.Pagination = nil
	successResp.Meta.Message = "Content created successfully"
	return ctx.Status(fiber.StatusCreated).JSON(successResp)
}

func (handler *contentHandler) EditContentByID(ctx *fiber.Ctx) error {
	userID, ok := getAuthenticatedUserID(ctx)
	if !ok {
		code := "[HANDLER] UpdateContent - 1"
		log.Errorw(code, "reason", "unauthorized")
		return respondError(ctx, fiber.StatusUnauthorized, "Unauthorized")
	}

	var contentRequest request.ContentRequest
	if err := ctx.BodyParser(&contentRequest); err != nil {
		code := "[HANDLER] UpdateContent - 2"
		log.Errorw(code, err)
		return respondError(ctx, fiber.StatusBadRequest, "Failed to parse request body")
	}

	if err := validation.ValidateStruct(&contentRequest); err != nil {
		code := "[HANDLER] UpdateContent - 3"
		log.Errorw(code, err)
		return respondError(ctx, fiber.StatusBadRequest, err.Error())
	}

	idParams := ctx.Params("id")
	id, err := conv.StringToInt64(idParams)
	if err != nil {
		code := "[HANDLER] UpdateContent - 4"
		log.Errorw(code, err)
		return respondError(ctx, fiber.StatusBadRequest, "Invalid ID format")
	}

	reqEntity := entity.ContentEntity{
		ID:          id,
		Title:       contentRequest.Title,
		Excerpt:     contentRequest.Excerpt,
		Description: contentRequest.Description,
		Image:       contentRequest.Image,
		Tags:        parseTags(contentRequest.Tags),
		Status:      contentRequest.Status,
		CategoryID:  contentRequest.CategoryID,
		CreatedByID: userID,
	}

	err = handler.contentService.UpdateContent(ctx.Context(), reqEntity)
	if err != nil {
		if errors.Is(err, contentError.ErrorContentNotFound) {
			code := "[HANDLER] UpdateContent - 5"
			log.Errorw(code, err)
			return respondError(ctx, fiber.StatusNotFound, "Content not found")
		}
		code := "[HANDLER] UpdateContent - 6"
		log.Errorw(code, err)
		return respondError(ctx, fiber.StatusInternalServerError, "Failed to update content")
	}

	var successResp response.DefaultSuccessResponse
	successResp.Meta.Status = true
	successResp.Pagination = nil
	successResp.Meta.Message = "Content updated successfully"
	return ctx.Status(fiber.StatusOK).JSON(successResp)
}

func (handler *contentHandler) DeleteContentByID(ctx *fiber.Ctx) error {
	if _, ok := getAuthenticatedUserID(ctx); !ok {
		code := "[HANDLER] DeleteContentByID - 1"
		log.Errorw(code, "reason", "unauthorized")
		return respondError(ctx, fiber.StatusUnauthorized, "Unauthorized")
	}

	idParams := ctx.Params("id")
	id, err := conv.StringToInt64(idParams)
	if err != nil {
		code := "[HANDLER] DeleteContentByID - 2"
		log.Errorw(code, err)
		return respondError(ctx, fiber.StatusBadRequest, "Invalid ID format")
	}

	err = handler.contentService.DeleteContentByID(ctx.Context(), id)
	if err != nil {
		if errors.Is(err, contentError.ErrorContentNotFound) {
			code := "[HANDLER] DeleteContentByID - 3"
			log.Errorw(code, err)
			return respondError(ctx, fiber.StatusNotFound, "Content not found")
		}
		code := "[HANDLER] DeleteContentByID - 4"
		log.Errorw(code, err)
		return respondError(ctx, fiber.StatusInternalServerError, "Failed to delete content")
	}

	var successResp response.DefaultSuccessResponse
	successResp.Meta.Status = true
	successResp.Pagination = nil
	successResp.Meta.Message = "Content deleted successfully"
	return ctx.Status(fiber.StatusOK).JSON(successResp)
}

func (handler *contentHandler) UploadImageR2(ctx *fiber.Ctx) error {
	userID, ok := getAuthenticatedUserID(ctx)
	if !ok {
		code := "[HANDLER] UploadImageR2 - 1"
		log.Errorw(code, "reason", "unauthorized")
		return respondError(ctx, fiber.StatusUnauthorized, "Unauthorized")
	}

	file, err := ctx.FormFile("image")
	if err != nil {
		code := "[HANDLER] UploadImageR2 - 2"
		log.Errorw(code, err)
		return respondError(ctx, fiber.StatusBadRequest, "Failed to parse request body")
	}

	if file.Size <= 0 || file.Size > maxImageUploadSize {
		return respondError(ctx, fiber.StatusBadRequest, "Image size must be between 1 byte and 5MB")
	}

	f, err := file.Open()
	if err != nil {
		code := "[HANDLER] UploadImageR2 - 3"
		log.Errorw(code, err)
		return respondError(ctx, fiber.StatusBadRequest, "Failed to open uploaded file")
	}
	defer f.Close()

	header := make([]byte, 512)
	n, err := io.ReadFull(f, header)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		code := "[HANDLER] UploadImageR2 - 4"
		log.Errorw(code, err)
		return respondError(ctx, fiber.StatusBadRequest, "Failed to read uploaded file")
	}

	detectedType := http.DetectContentType(header[:n])
	ext, allowed := allowedImageTypes[detectedType]
	if !allowed {
		return respondError(ctx, fiber.StatusBadRequest, "Unsupported image type; allowed types: jpeg, png, webp, gif")
	}

	remainingAllowed := int64(maxImageUploadSize - n)
	body := io.MultiReader(bytes.NewReader(header[:n]), io.LimitReader(f, remainingAllowed))
	fileName := fmt.Sprintf("%d-%d-%s%s", userID, time.Now().UnixNano(), uuid.NewString(), ext)

	imageURL, err := handler.contentService.UploadImageR2(ctx.Context(), entity.FileUploadEntity{
		Name:        fileName,
		Size:        file.Size,
		ContentType: detectedType,
	}, body)

	if err != nil {
		code := "[HANDLER] UploadImageR2 - 5"
		log.Errorw(code, err)
		return respondError(ctx, fiber.StatusInternalServerError, "Failed to upload image")
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
	reqEntity, err := parseContentQuery(ctx, 6, true)
	if err != nil {
		code := "[HANDLER] GetContentWithQuery - 1"
		log.Errorw(code, err)
		return respondError(ctx, fiber.StatusBadRequest, err.Error())
	}

	results, totalData, totalPages, err := handler.contentService.GetContents(ctx.Context(), reqEntity)
	if err != nil {
		code := "[HANDLER] GetContentWithQuery - 2"
		log.Errorw(code, err)
		return respondError(ctx, fiber.StatusInternalServerError, "Failed to fetch contents")
	}

	responses := make([]response.ContentResponse, 0, len(results))
	for _, result := range results {
		responses = append(responses, mapContentResponse(result))
	}

	var successResp response.DefaultSuccessResponse
	successResp.Meta.Status = true
	successResp.Pagination = &response.PaginationResponse{
		TotalRecord: totalData,
		TotalPage:   int64(totalPages),
		Page:        int64(reqEntity.Page),
		PerPage:     int64(reqEntity.Limit),
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
		return respondError(ctx, fiber.StatusBadRequest, "Invalid ID format")
	}

	result, err := handler.contentService.GetContentByID(ctx.Context(), id)
	if err != nil {
		if errors.Is(err, contentError.ErrorContentNotFound) {
			code := "[HANDLER] GetContentDetail - 2"
			log.Errorw(code, err)
			return respondError(ctx, fiber.StatusNotFound, "Content not found")
		}

		code := "[HANDLER] GetContentDetail - 3"
		log.Errorw(code, err)
		return respondError(ctx, fiber.StatusInternalServerError, "Failed to get content")
	}

	if result.Status != "PUBLISH" {
		code := "[HANDLER] GetContentDetail - 4"
		log.Errorw(code, "reason", "content not published")
		return respondError(ctx, fiber.StatusNotFound, "Content not found")
	}

	var successResp response.DefaultSuccessResponse
	successResp.Meta.Status = true
	successResp.Pagination = nil
	successResp.Meta.Message = "Content fetched successfully"
	successResp.Data = mapContentResponse(*result)
	return ctx.JSON(successResp)
}
