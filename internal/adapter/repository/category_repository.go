package repository

import (
	"context"
	"errors"

	"github.com/JonathanGunawan30/portal-news-backend.git/internal/core/domain/entity"
	categoryError "github.com/JonathanGunawan30/portal-news-backend.git/internal/core/domain/errors"
	"github.com/JonathanGunawan30/portal-news-backend.git/internal/core/domain/model"
	"github.com/JonathanGunawan30/portal-news-backend.git/lib/conv"
	"github.com/gofiber/fiber/v2/log"
	"gorm.io/gorm"
)

type CategoryRepository interface {
	GetCategories(ctx context.Context) ([]entity.CategoryEntity, error)
	GetCategoryByID(ctx context.Context, id int64) (*entity.CategoryEntity, error)
	CreateCategory(ctx context.Context, request entity.CategoryEntity) error
	EditCategoryByID(ctx context.Context, request entity.CategoryEntity) error
	DeleteCategoryByID(ctx context.Context, id int64) error
	IsTitleExists(ctx context.Context, title string) (bool, error)
}

type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{
		db: db,
	}
}

func (repository *categoryRepository) GetCategories(ctx context.Context) ([]entity.CategoryEntity, error) {
	var modelCategories []model.Category

	err := repository.db.WithContext(ctx).Order("created_at DESC").Preload("User").Find(&modelCategories).Error

	if err != nil {
		code := "[REPOSITORY] GetCategories - 1"
		log.Errorw(code, err)
		return nil, err
	}

	if len(modelCategories) == 0 {
		return []entity.CategoryEntity{}, nil
	}

	var responses []entity.CategoryEntity

	for _, response := range modelCategories {
		responses = append(responses, entity.CategoryEntity{
			ID:    response.ID,
			Title: response.Title,
			Slug:  response.Slug,
			User: entity.UserEntity{
				ID:    response.User.ID,
				Name:  response.User.Name,
				Email: response.User.Email,
			},
		})
	}

	return responses, nil

}

func (repository *categoryRepository) GetCategoryByID(ctx context.Context, id int64) (*entity.CategoryEntity, error) {
	var modelCategory model.Category
	result := repository.db.WithContext(ctx).Preload("User").Take(&modelCategory, id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, categoryError.ErrorCategoryNotFound
		}
		code := "[REPOSITORY] GetCategoryByID - 1"
		log.Errorw(code, result.Error)
		return nil, result.Error
	}

	return &entity.CategoryEntity{
		ID:    modelCategory.ID,
		Title: modelCategory.Title,
		Slug:  modelCategory.Slug,
		User: entity.UserEntity{
			ID:    modelCategory.User.ID,
			Name:  modelCategory.User.Name,
			Email: modelCategory.User.Email,
		},
	}, nil
}

func (repository *categoryRepository) CreateCategory(ctx context.Context, request entity.CategoryEntity) error {
	modelCategory := model.Category{
		Title:       request.Title,
		Slug:        request.Slug,
		CreatedByID: request.User.ID,
	}

	err := repository.db.WithContext(ctx).Create(&modelCategory).Error
	if err != nil {
		code := "[REPOSITORY] CreateCategory - 1"
		log.Errorw(code, err)
		return err
	}
	return nil
}

func (repository *categoryRepository) EditCategoryByID(ctx context.Context, request entity.CategoryEntity) error {
	var foundCategory model.Category

	result := repository.db.WithContext(ctx).
		Where("(title = ? OR slug = ?) AND id <> ?", request.Title, request.Slug, request.ID).
		First(&foundCategory)

	if result.Error == nil {
		return categoryError.ErrorCategoryTitleAlreadyExist
	} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		code := "[REPOSITORY] EditCategoryByID - 1"
		log.Errorw(code, result.Error)
		return result.Error
	}

	updateData := model.Category{
		Title: request.Title,
		Slug:  request.Slug,
	}

	updateResult := repository.db.WithContext(ctx).
		Model(&model.Category{}).
		Where("id = ?", request.ID).
		Updates(updateData)

	if updateResult.Error != nil {
		code := "[REPOSITORY] EditCategoryByID - 2"
		log.Errorw(code, updateResult.Error)
		return updateResult.Error
	}
	if updateResult.RowsAffected == 0 {
		return categoryError.ErrorCategoryNotFoundToUpdate
	}
	return nil
}

func (repository *categoryRepository) DeleteCategoryByID(ctx context.Context, id int64) error {
	var count int64
	err := repository.db.WithContext(ctx).Model(&model.Content{}).Where("category_id = ?", id).Count(&count).Error
	if err != nil {
		code := "[REPOSITORY] DeleteCategoryByID - 1"
		log.Errorw(code, err)
		return err
	}

	if count > 0 {
		return categoryError.ErrorCategoryHasAssociatedContent
	}

	result := repository.db.WithContext(ctx).Delete(&model.Category{}, id)
	if result.Error != nil {
		code := "[REPOSITORY] DeleteCategoryByID - 2"
		log.Errorw(code, result.Error)
		return result.Error
	}

	if result.RowsAffected == 0 {
		code := "[REPOSITORY] DeleteCategoryByID - 3"
		log.Errorw(code, "reason", "category not found")
		return categoryError.ErrorCategoryNotFoundToDelete
	}

	return nil
}

func (repository *categoryRepository) IsTitleExists(ctx context.Context, title string) (bool, error) {
	var count int64
	slug := conv.GenerateSlug(title)
	err := repository.db.WithContext(ctx).Model(model.Category{}).Where("title = ? OR slug = ?", title, slug).Count(&count).Error
	if err != nil {
		code := "[REPOSITORY] IsTitleExist - 1"
		log.Errorw(code, err)
		return false, err
	}
	return count > 0, nil
}
