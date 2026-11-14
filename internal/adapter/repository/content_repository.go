package repository

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/JonathanGunawan30/portal-news-backend.git/internal/core/domain/entity"
	contentError "github.com/JonathanGunawan30/portal-news-backend.git/internal/core/domain/errors"
	"github.com/JonathanGunawan30/portal-news-backend.git/internal/core/domain/model"
	"github.com/gofiber/fiber/v2/log"
	"gorm.io/gorm"
)

var allowedContentOrderColumns = map[string]string{
	"id":         "id",
	"title":      "title",
	"created_at": "created_at",
	"updated_at": "updated_at",
}

type ContentRepository interface {
	GetContents(ctx context.Context, query entity.QueryString) ([]entity.ContentEntity, int64, int64, error)
	GetContentByID(ctx context.Context, id int64) (*entity.ContentEntity, error)
	CreateContent(ctx context.Context, request entity.ContentEntity) error
	UpdateContent(ctx context.Context, request entity.ContentEntity) error
	DeleteContentByID(ctx context.Context, id int64) error
}

type contentRepository struct {
	db *gorm.DB
}

func NewContentRepository(db *gorm.DB) ContentRepository {
	return &contentRepository{
		db: db,
	}
}

func sanitizeOrderClause(orderBy, orderType string) string {
	col, ok := allowedContentOrderColumns[strings.ToLower(strings.TrimSpace(orderBy))]
	if !ok {
		col = "created_at"
	}

	dir := "DESC"
	if strings.EqualFold(strings.TrimSpace(orderType), "ASC") {
		dir = "ASC"
	}

	return fmt.Sprintf("%s %s", col, dir)
}

func splitTags(raw string) []string {
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

func (repository *contentRepository) GetContents(ctx context.Context, query entity.QueryString) ([]entity.ContentEntity, int64, int64, error) {
	var modelContents []model.Content
	var totalRows int64

	page := query.Page
	if page <= 0 {
		page = 1
	}

	limit := query.Limit
	if limit <= 0 {
		limit = 10
	} else if limit > 100 {
		limit = 100
	}

	order := sanitizeOrderClause(query.OrderBy, query.OrderType)
	offset := (page - 1) * limit

	queryBuilder := repository.db.WithContext(ctx).Model(&model.Content{})
	if query.Search != "" {
		searchQuery := "%" + query.Search + "%"
		queryBuilder = queryBuilder.Where("title ILIKE ? OR excerpt ILIKE ? OR description ILIKE ?", searchQuery, searchQuery, searchQuery)
	}
	if query.CategoryID > 0 {
		queryBuilder = queryBuilder.Where("category_id = ?", query.CategoryID)
	}
	if query.Status != "" {
		queryBuilder = queryBuilder.Where("status = ?", query.Status)
	}

	err := queryBuilder.Count(&totalRows).Error
	if err != nil {
		code := "[REPOSITORY] GetContents - Count"
		log.Errorw(code, err)
		return nil, 0, 0, err
	}

	if totalRows == 0 {
		return []entity.ContentEntity{}, 0, 0, nil
	}

	err = queryBuilder.
		Order(order).
		Limit(limit).
		Offset(offset).
		Preload("User").
		Preload("Category").
		Find(&modelContents).Error

	if err != nil {
		code := "[REPOSITORY] GetContents - Find"
		log.Errorw(code, err)
		return nil, 0, 0, err
	}

	var responses []entity.ContentEntity
	for _, content := range modelContents {
		response := entity.ContentEntity{
			ID:          content.ID,
			Title:       content.Title,
			Excerpt:     content.Excerpt,
			Description: content.Description,
			Image:       content.Image,
			Tags:        splitTags(content.Tags),
			Status:      content.Status,
			CategoryID:  content.CategoryID,
			CreatedByID: content.CreatedByID,
			CreatedAt:   content.CreatedAt,
			UpdatedAt:   content.UpdatedAt,
			Category: entity.CategoryEntity{
				ID:    content.Category.ID,
				Title: content.Category.Title,
				Slug:  content.Category.Slug,
			},
			User: entity.UserEntity{
				ID:    content.User.ID,
				Name:  content.User.Name,
				Email: content.User.Email,
			},
		}
		responses = append(responses, response)
	}
	totalPages := int(math.Ceil(float64(totalRows) / float64(limit)))

	return responses, totalRows, int64(totalPages), nil
}

func (repository *contentRepository) GetContentByID(ctx context.Context, id int64) (*entity.ContentEntity, error) {
	var modelContent model.Content
	result := repository.db.WithContext(ctx).Preload("User").Preload("Category").Take(&modelContent, id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, contentError.ErrorContentNotFound
		}
		code := "[REPOSITORY] GetContentByID - 1"
		log.Errorw(code, result.Error)
		return nil, result.Error
	}

	response := entity.ContentEntity{
		ID:          modelContent.ID,
		Title:       modelContent.Title,
		Excerpt:     modelContent.Excerpt,
		Description: modelContent.Description,
		Image:       modelContent.Image,
		Tags:        splitTags(modelContent.Tags),
		Status:      modelContent.Status,
		CategoryID:  modelContent.CategoryID,
		CreatedByID: modelContent.CreatedByID,
		CreatedAt:   modelContent.CreatedAt,
		UpdatedAt:   modelContent.UpdatedAt,
		Category: entity.CategoryEntity{
			ID:    modelContent.Category.ID,
			Title: modelContent.Category.Title,
			Slug:  modelContent.Category.Slug,
		},
		User: entity.UserEntity{
			ID:    modelContent.User.ID,
			Name:  modelContent.User.Name,
			Email: modelContent.User.Email,
		},
	}
	return &response, nil

}

func (repository *contentRepository) CreateContent(ctx context.Context, request entity.ContentEntity) error {
	tags := strings.Join(request.Tags, ",")
	modelContent := model.Content{
		Title:       request.Title,
		Excerpt:     request.Excerpt,
		Description: request.Description,
		Image:       request.Image,
		Tags:        tags,
		Status:      request.Status,
		CategoryID:  request.CategoryID,
		CreatedByID: request.CreatedByID,
	}
	err := repository.db.WithContext(ctx).Create(&modelContent).Error
	if err != nil {
		code := "[REPOSITORY] CreateContent - 1"
		log.Errorw(code, err)
		return err
	}
	return nil
}

func (repository *contentRepository) UpdateContent(ctx context.Context, request entity.ContentEntity) error {
	tags := strings.Join(request.Tags, ",")
	updates := map[string]any{
		"title":       request.Title,
		"excerpt":     request.Excerpt,
		"description": request.Description,
		"image":       request.Image,
		"tags":        tags,
		"category_id": request.CategoryID,
	}
	if request.Status != "" {
		updates["status"] = request.Status
	}

	result := repository.db.WithContext(ctx).Model(&model.Content{}).Where("id = ?", request.ID).Updates(updates)

	if result.Error != nil {
		code := "[REPOSITORY] UpdateContent - 1"
		log.Errorw(code, result.Error)
		return result.Error
	}

	if result.RowsAffected == 0 {
		code := "[REPOSITORY] UpdateContent - 2"
		log.Errorw(code, "reason", "content not found")
		return contentError.ErrorContentNotFound
	}

	return nil
}

func (repository *contentRepository) DeleteContentByID(ctx context.Context, id int64) error {
	result := repository.db.WithContext(ctx).Delete(&model.Content{}, id)

	if result.Error != nil {
		code := "[REPOSITORY] DeleteContentByID - 1"
		log.Errorw(code, result.Error)
		return result.Error
	}

	if result.RowsAffected == 0 {
		code := "[REPOSITORY] DeleteContentByID - 2"
		log.Errorw(code, "reason", "content not found")
		return contentError.ErrorContentNotFound
	}

	return nil
}
