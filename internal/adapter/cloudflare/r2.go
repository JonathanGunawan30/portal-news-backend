package cloudflare

import (
	"context"
	"fmt"
	"os"

	"github.com/JonathanGunawan30/portal-news-backend.git/config"
	"github.com/JonathanGunawan30/portal-news-backend.git/internal/core/domain/entity"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gofiber/fiber/v2/log"
)

type CloudflareR2Adapter interface {
	UploadImage(request *entity.FileUploadEntity) (string, error)
}

type cloudflareR2Adapter struct {
	Client  *s3.Client
	Bucket  string
	BaseURL string
}

func NewCloudflareR2Adapter(client *s3.Client, cfg *config.Config) CloudflareR2Adapter {
	clientBase := s3.NewFromConfig(cfg.LoadAwsConfig(), func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.R2.AccountID))
	})
	return &cloudflareR2Adapter{
		Client:  clientBase,
		Bucket:  cfg.R2.Name,
		BaseURL: cfg.R2.PublicUrl,
	}
}

func (cloudflare *cloudflareR2Adapter) UploadImage(request *entity.FileUploadEntity) (string, error) {
	openedFile, err := os.Open(request.Path)
	if err != nil {
		code := "[CLOUDFLARE R2] UploadImage - 1"
		log.Errorw(code, err)
		return "", err
	}

	defer openedFile.Close()

	_, err = cloudflare.Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(cloudflare.Bucket),
		Key:         aws.String(request.Name),
		Body:        openedFile,
		ContentType: aws.String("image/jpeg"),
	})

	if err != nil {
		code := "[CLOUDFLARE R2] UploadImage - 2"
		log.Errorw(code, err)
		return "", err
	}

	return fmt.Sprintf("%s/%s", cloudflare.BaseURL, request.Name), nil

}
