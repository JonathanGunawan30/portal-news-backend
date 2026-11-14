package service

import (
	"context"
	"io"

	"github.com/JonathanGunawan30/portal-news-backend.git/config"
	"github.com/JonathanGunawan30/portal-news-backend.git/internal/adapter/cloudflare"
	"github.com/JonathanGunawan30/portal-news-backend.git/internal/adapter/repository"
	"github.com/JonathanGunawan30/portal-news-backend.git/internal/core/domain/entity"
	"github.com/gofiber/fiber/v2/log"
)

type ContentService interface {
	GetContents(ctx context.Context, query entity.QueryString) ([]entity.ContentEntity, int64, int, error)
	GetContentByID(ctx context.Context, id int64) (*entity.ContentEntity, error)
	CreateContent(ctx context.Context, request entity.ContentEntity) error
	UpdateContent(ctx context.Context, request entity.ContentEntity) error
	DeleteContentByID(ctx context.Context, id int64) error
	UploadImageR2(ctx context.Context, req entity.FileUploadEntity, body io.Reader) (string, error)
}

type contentService struct {
	contentRepository repository.ContentRepository
	cfg               *config.Config
	r2                cloudflare.CloudflareR2Adapter
}

func NewContentService(contentRepository repository.ContentRepository, cfg *config.Config, r2 cloudflare.CloudflareR2Adapter) ContentService {
	return &contentService{
		contentRepository: contentRepository,
		cfg:               cfg,
		r2:                r2,
	}
}

func (service *contentService) GetContents(ctx context.Context, query entity.QueryString) ([]entity.ContentEntity, int64, int, error) {
	results, totalData, totalPages, err := service.contentRepository.GetContents(ctx, query)
	if err != nil {
		code := "[SERVICE] GetContents - 1"
		log.Errorw(code, err)
		return nil, 0, 0, err
	}

	return results, totalData, int(totalPages), nil
}

func (service *contentService) GetContentByID(ctx context.Context, id int64) (*entity.ContentEntity, error) {
	result, err := service.contentRepository.GetContentByID(ctx, id)
	if err != nil {
		code := "[SERVICE] GetContentByID - 1"
		log.Errorw(code, err)
		return nil, err
	}

	return result, nil
}

func (service *contentService) CreateContent(ctx context.Context, request entity.ContentEntity) error {
	err := service.contentRepository.CreateContent(ctx, request)
	if err != nil {
		code := "[SERVICE] CreateContent - 1"
		log.Errorw(code, err)
		return err
	}
	return nil
}

func (service *contentService) UpdateContent(ctx context.Context, request entity.ContentEntity) error {
	err := service.contentRepository.UpdateContent(ctx, request)
	if err != nil {
		code := "[SERVICE] UpdateContent - 1"
		log.Errorw(code, err)
		return err
	}
	return nil
}

func (service *contentService) DeleteContentByID(ctx context.Context, id int64) error {
	err := service.contentRepository.DeleteContentByID(ctx, id)
	if err != nil {
		code := "[SERVICE] DeleteContentByID - 1"
		log.Errorw(code, err)
		return err
	}
	return nil
}

func (service *contentService) UploadImageR2(ctx context.Context, req entity.FileUploadEntity, body io.Reader) (string, error) {
	url, err := service.r2.UploadImage(ctx, req, body)
	if err != nil {
		code := "[SERVICE] UploadImageR2 - 1"
		log.Errorw(code, err)
		return "", err
	}
	return url, nil
}
