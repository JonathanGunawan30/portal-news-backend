package handler

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JonathanGunawan30/portal-news-backend.git/config"
	"github.com/JonathanGunawan30/portal-news-backend.git/internal/core/domain/entity"
	domainErr "github.com/JonathanGunawan30/portal-news-backend.git/internal/core/domain/errors"
	"github.com/JonathanGunawan30/portal-news-backend.git/lib/auth"
	"github.com/JonathanGunawan30/portal-news-backend.git/lib/middleware"
	"github.com/gofiber/fiber/v2"
)

type stubContentService struct{}

func (s *stubContentService) GetContents(context.Context, entity.QueryString) ([]entity.ContentEntity, int64, int, error) {
	return []entity.ContentEntity{}, 0, 0, nil
}
func (s *stubContentService) GetContentByID(context.Context, int64) (*entity.ContentEntity, error) {
	return nil, domainErr.ErrorContentNotFound
}
func (s *stubContentService) CreateContent(context.Context, entity.ContentEntity) error { return nil }
func (s *stubContentService) UpdateContent(context.Context, entity.ContentEntity) error { return nil }
func (s *stubContentService) DeleteContentByID(context.Context, int64) error            { return nil }
func (s *stubContentService) UploadImageR2(_ context.Context, req entity.FileUploadEntity, _ io.Reader) (string, error) {
	return "https://cdn.example.com/" + req.Name, nil
}

type stubUserService struct{}

func (s *stubUserService) UpdatePassword(context.Context, int64, string, string) error {
	return domainErr.ErrorUserNotFound
}
func (s *stubUserService) GetUserByID(context.Context, int64) (*entity.UserEntity, error) {
	return nil, domainErr.ErrorUserNotFound
}

func TestParseContentQuerySecurityAndOrderFix(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		q, err := parseContentQuery(c, 6, true)
		if err != nil {
			return respondError(c, fiber.StatusBadRequest, err.Error())
		}
		return c.JSON(q)
	})

	// Valid ?order=title (previously broken by order vs orderBy mismatch)
	req := httptest.NewRequest(http.MethodGet, "/test?order=title&orderType=asc&limit=500", nil)
	resp, err := app.Test(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK for valid query, got status=%d err=%v", resp.StatusCode, err)
	}

	// SQLi attempt in order param must be rejected with 400
	reqSQLi := httptest.NewRequest(http.MethodGet, "/test?order=created_at;DROP%20TABLE%20contents", nil)
	respSQLi, err := app.Test(reqSQLi)
	if err != nil || respSQLi.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 BadRequest for SQLi order param, got status=%d err=%v", respSQLi.StatusCode, err)
	}

	// Negative page must be rejected with 400
	reqPage := httptest.NewRequest(http.MethodGet, "/test?page=0", nil)
	respPage, err := app.Test(reqPage)
	if err != nil || respPage.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 BadRequest for page=0, got status=%d err=%v", respPage.StatusCode, err)
	}
}

func TestAuthenticatedUserIDNoNilPanic(t *testing.T) {
	app := fiber.New()
	app.Get("/profile", func(c *fiber.Ctx) error {
		if _, ok := getAuthenticatedUserID(c); !ok {
			return respondError(c, fiber.StatusUnauthorized, "Unauthorized")
		}
		return c.SendStatus(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/profile", nil)
	resp, err := app.Test(req)
	if err != nil || resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 without panic, got status=%d err=%v", resp.StatusCode, err)
	}
}

func TestUserNotFoundReturns404(t *testing.T) {
	cfg := &config.Config{App: config.App{JwtSecretKey: "super-secret-key-at-least-32-chars!!"}}
	jwtAuth := auth.NewJwt(cfg)
	h := NewUserHandler(&stubUserService{}, cfg, jwtAuth)

	app := fiber.New()
	app.Get("/profile", func(c *fiber.Ctx) error {
		c.Locals("user", &entity.JwtData{UserID: 999})
		return h.GetUserByID(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/profile", nil)
	resp, err := app.Test(req)
	if err != nil || resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 NotFound, got status=%d err=%v", resp.StatusCode, err)
	}
}

func TestUploadImageValidationAndOriginCheck(t *testing.T) {
	cfg := &config.Config{App: config.App{JwtSecretKey: "super-secret-key-at-least-32-chars!!"}}
	mw := middleware.NewMiddleware(auth.NewJwt(cfg))
	h := NewContentHandler(&stubContentService{})

	app := fiber.New()
	app.Use(mw.CheckOrigin("https://blog.jonathangunawan.com"))
	app.Post("/upload", func(c *fiber.Ctx) error {
		c.Locals("user", &entity.JwtData{UserID: 1})
		return h.UploadImageR2(c)
	})

	makeMultipart := func(filename string, content []byte) (*bytes.Buffer, string) {
		var buf bytes.Buffer
		w := multipart.NewWriter(&buf)
		fw, _ := w.CreateFormFile("image", filename)
		_, _ = fw.Write(content)
		_ = w.Close()
		return &buf, w.FormDataContentType()
	}

	// Disguised HTML file must be rejected
	htmlBuf, htmlCT := makeMultipart("evil.jpg", []byte("<html><script>alert(1)</script></html>"))
	reqHTML := httptest.NewRequest(http.MethodPost, "/upload", htmlBuf)
	reqHTML.Header.Set("Content-Type", htmlCT)
	reqHTML.Header.Set("Origin", "https://blog.jonathangunawan.com")
	respHTML, err := app.Test(reqHTML)
	if err != nil || respHTML.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for disguised HTML upload, got status=%d err=%v", respHTML.StatusCode, err)
	}

	// Foreign browser origin must be rejected with 403
	pngBytes := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D}
	pngBuf, pngCT := makeMultipart("valid.png", pngBytes)
	reqEvilOrigin := httptest.NewRequest(http.MethodPost, "/upload", pngBuf)
	reqEvilOrigin.Header.Set("Content-Type", pngCT)
	reqEvilOrigin.Header.Set("Origin", "https://evil.example.com")
	respEvil, err := app.Test(reqEvilOrigin)
	if err != nil || respEvil.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 for untrusted origin, got status=%d err=%v", respEvil.StatusCode, err)
	}

	// Valid PNG from allowed origin succeeds
	pngBuf2, pngCT2 := makeMultipart("../../weird.png", pngBytes)
	reqValid := httptest.NewRequest(http.MethodPost, "/upload", pngBuf2)
	reqValid.Header.Set("Content-Type", pngCT2)
	reqValid.Header.Set("Origin", "https://blog.jonathangunawan.com")
	respValid, err := app.Test(reqValid)
	if err != nil || respValid.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 for valid PNG upload, got status=%d err=%v", respValid.StatusCode, err)
	}
}
