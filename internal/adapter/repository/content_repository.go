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

func (repository *contentRepository) GetContents(ctx context.Context, query entity.QueryString) ([]entity.ContentEntity, int64, int64, error) {
	var modelContents []model.Content
	var totalRows int64

	order := fmt.Sprintf("%s %s", query.OrderBy, query.OrderType)
	offset := (query.Page - 1) * query.Limit

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
		Limit(query.Limit).
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
		tags := strings.Split(content.Tags, ",")
		response := entity.ContentEntity{
			ID:          content.ID,
			Title:       content.Title,
			Excerpt:     content.Excerpt,
			Description: content.Description,
			Image:       content.Image,
			Tags:        tags,
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
	totalPages := int(math.Ceil(float64(totalRows) / float64(query.Limit)))

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

	tags := strings.Split(modelContent.Tags, ",")
	response := entity.ContentEntity{
		ID:          modelContent.ID,
		Title:       modelContent.Title,
		Excerpt:     modelContent.Excerpt,
		Description: modelContent.Description,
		Image:       modelContent.Image,
		Tags:        tags,
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

	result := repository.db.WithContext(ctx).Where("id = ?", request.ID).Updates(&modelContent)

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
