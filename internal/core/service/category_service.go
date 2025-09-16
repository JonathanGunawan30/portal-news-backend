package service

import (
	"context"

	"github.com/JonathanGunawan30/portal-news-backend.git/internal/adapter/repository"
	"github.com/JonathanGunawan30/portal-news-backend.git/internal/core/domain/entity"
	categoryError "github.com/JonathanGunawan30/portal-news-backend.git/internal/core/domain/errors"
	"github.com/JonathanGunawan30/portal-news-backend.git/lib/conv"
	"github.com/gofiber/fiber/v2/log"
)

type CategoryService interface {
	GetCategories(ctx context.Context) ([]entity.CategoryEntity, error)
	GetCategoryByID(ctx context.Context, id int64) (*entity.CategoryEntity, error)
	CreateCategory(ctx context.Context, request entity.CategoryEntity) error
	EditCategoryByID(ctx context.Context, requests entity.CategoryEntity) error
	DeleteCategoryByID(ctx context.Context, id int64) error
}

func NewCategoryService(categoryRepository repository.CategoryRepository) CategoryService {
	return &categoryService{
		categoryRepository: categoryRepository,
	}
}

type categoryService struct {
	categoryRepository repository.CategoryRepository
}

func (service *categoryService) GetCategories(ctx context.Context) ([]entity.CategoryEntity, error) {
	results, err := service.categoryRepository.GetCategories(ctx)
	if err != nil {
		code := "[SERVICE] GetCategories - 1"
		log.Errorw(code, err)
		return nil, err
	}

	return results, nil
}

func (service *categoryService) GetCategoryByID(ctx context.Context, id int64) (*entity.CategoryEntity, error) {
	categoryEntity, err := service.categoryRepository.GetCategoryByID(ctx, id)
	if err != nil {
		code := "[SERVICE] GetCategoryByID - 1"
		log.Errorw(code, err)
		return nil, err
	}

	return categoryEntity, nil
}

func (service *categoryService) CreateCategory(ctx context.Context, request entity.CategoryEntity) error {
	isTitleExists, err := service.categoryRepository.IsTitleExists(ctx, request.Title)
	if err != nil {
		code := "[SERVICE] CreateCategory - 1"
		log.Errorw(code, err)
		return err
	}

	if isTitleExists {
		return categoryError.ErrorCategoryTitleAlreadyExist
	}

	slug := conv.GenerateSlug(request.Title)
	request.Slug = slug

	err = service.categoryRepository.CreateCategory(ctx, request)
	if err != nil {
		code := "[SERVICE] CreateCategory - 2"
		log.Errorw(code, err)
		return err
	}
	return nil
}

func (service *categoryService) EditCategoryByID(ctx context.Context, request entity.CategoryEntity) error {
	slug := conv.GenerateSlug(request.Title)
	request.Slug = slug

	err := service.categoryRepository.EditCategoryByID(ctx, request)
	if err != nil {
		code := "[SERVICE] EditCategoryByID - 1"
		log.Errorw(code, err)
		return err
	}
	return nil
}

func (service *categoryService) DeleteCategoryByID(ctx context.Context, id int64) error {
	err := service.categoryRepository.DeleteCategoryByID(ctx, id)
	if err != nil {
		code := "[SERVICE] DeleteCategoryByID - 1"
		log.Errorw(code, err)
		return err
	}
	return nil
}
