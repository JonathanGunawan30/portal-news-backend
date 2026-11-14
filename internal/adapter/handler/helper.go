package handler

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/JonathanGunawan30/portal-news-backend.git/config"
	"github.com/JonathanGunawan30/portal-news-backend.git/internal/adapter/handler/response"
	"github.com/JonathanGunawan30/portal-news-backend.git/internal/core/domain/entity"
	"github.com/JonathanGunawan30/portal-news-backend.git/lib/conv"
	"github.com/gofiber/fiber/v2"
)

var allowedOrderColumns = map[string]string{
	"id":         "id",
	"title":      "title",
	"created_at": "created_at",
	"updated_at": "updated_at",
}

func respondError(ctx *fiber.Ctx, status int, message string) error {
	var resp response.ErrorResponseDefault
	resp.Meta.Status = false
	resp.Meta.Message = message
	return ctx.Status(status).JSON(resp)
}

func getAuthenticatedUserID(ctx *fiber.Ctx) (int64, bool) {
	claims, ok := ctx.Locals("user").(*entity.JwtData)
	if !ok || claims == nil || claims.UserID <= 0 {
		return 0, false
	}
	return claims.UserID, true
}

func isProductionEnv(cfg *config.Config) bool {
	if cfg != nil && strings.TrimSpace(cfg.App.AppEnv) != "" {
		return strings.EqualFold(strings.TrimSpace(cfg.App.AppEnv), "production")
	}
	return strings.EqualFold(strings.TrimSpace(os.Getenv("APP_ENV")), "production")
}

func setAuthCookie(ctx *fiber.Ctx, value string, expires time.Time, production bool) {
	sameSite := "Lax"
	secure := false
	if production {
		sameSite = "None"
		secure = true
	}

	ctx.Cookie(&fiber.Cookie{
		Name:     "X-API-TOKEN",
		Value:    value,
		Expires:  expires,
		HTTPOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		Path:     "/",
	})
}

func parseContentQuery(ctx *fiber.Ctx, defaultLimit int, publicOnly bool) (entity.QueryString, error) {
	page := 1
	if ctx.Query("page") != "" {
		parsedPage, err := conv.StringToInt(ctx.Query("page"))
		if err != nil || parsedPage < 1 {
			return entity.QueryString{}, errors.New("Invalid page format")
		}
		page = parsedPage
	}

	limit := defaultLimit
	if ctx.Query("limit") != "" {
		parsedLimit, err := conv.StringToInt(ctx.Query("limit"))
		if err != nil || parsedLimit < 1 {
			return entity.QueryString{}, errors.New("Invalid limit format")
		}
		if parsedLimit > 100 {
			parsedLimit = 100
		}
		limit = parsedLimit
	}

	order := "created_at"
	rawOrder := strings.TrimSpace(ctx.Query("order"))
	if rawOrder == "" {
		rawOrder = strings.TrimSpace(ctx.Query("orderBy"))
	}
	if rawOrder != "" {
		mappedCol, ok := allowedOrderColumns[strings.ToLower(rawOrder)]
		if !ok {
			return entity.QueryString{}, errors.New("Invalid order parameter")
		}
		order = mappedCol
	}

	orderType := "DESC"
	if rawOrderType := strings.TrimSpace(ctx.Query("orderType")); rawOrderType != "" {
		switch strings.ToUpper(rawOrderType) {
		case "ASC":
			orderType = "ASC"
		case "DESC":
			orderType = "DESC"
		default:
			return entity.QueryString{}, errors.New("Invalid orderType parameter")
		}
	}

	var categoryID int64
	if ctx.Query("category_id") != "" {
		parsedCatID, err := conv.StringToInt64(ctx.Query("category_id"))
		if err != nil || parsedCatID < 0 {
			return entity.QueryString{}, errors.New("Invalid category_id format")
		}
		categoryID = parsedCatID
	}

	status := ""
	if publicOnly {
		status = "PUBLISH"
	}

	return entity.QueryString{
		Limit:      limit,
		Page:       page,
		OrderBy:    order,
		OrderType:  orderType,
		Search:     strings.TrimSpace(ctx.Query("search")),
		CategoryID: categoryID,
		Status:     status,
	}, nil
}
