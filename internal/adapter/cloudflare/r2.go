package cloudflare

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/JonathanGunawan30/portal-news-backend.git/config"
	"github.com/JonathanGunawan30/portal-news-backend.git/internal/core/domain/entity"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gofiber/fiber/v2/log"
)

type CloudflareR2Adapter interface {
	UploadImage(ctx context.Context, request entity.FileUploadEntity, body io.Reader) (string, error)
}

type cloudflareR2Adapter struct {
	Client  *s3.Client
	Bucket  string
	BaseURL string
}

func NewCloudflareR2Adapter(cfg *config.Config) CloudflareR2Adapter {
	clientBase := s3.NewFromConfig(cfg.LoadAwsConfig(), func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.R2.AccountID))
	})
	return &cloudflareR2Adapter{
		Client:  clientBase,
		Bucket:  cfg.R2.Name,
		BaseURL: strings.TrimRight(cfg.R2.PublicUrl, "/"),
	}
}

func (cloudflare *cloudflareR2Adapter) UploadImage(ctx context.Context, request entity.FileUploadEntity, body io.Reader) (string, error) {
	uploadCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	contentType := request.ContentType
	if contentType == "" {
		contentType = "image/jpeg"
	}

	_, err := cloudflare.Client.PutObject(uploadCtx, &s3.PutObjectInput{
		Bucket:      aws.String(cloudflare.Bucket),
		Key:         aws.String(request.Name),
		Body:        body,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		code := "[CLOUDFLARE R2] UploadImage - 1"
		log.Errorw(code, err)
		return "", err
	}

	return fmt.Sprintf("%s/%s", cloudflare.BaseURL, request.Name), nil
}
