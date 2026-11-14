package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/JonathanGunawan30/portal-news-backend.git/internal/adapter/handler/response"
	"github.com/JonathanGunawan30/portal-news-backend.git/lib/auth"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type Middleware interface {
	CheckToken() fiber.Handler
	CheckOrigin(allowedOrigins string) fiber.Handler
}

type Options struct {
	authJwt auth.Jwt
}

func NewMiddleware(authJwt auth.Jwt) Middleware {
	return &Options{authJwt: authJwt}
}

func (o *Options) CheckToken() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var errorResponse response.ErrorResponseDefault

		tokenString := ctx.Cookies("X-API-TOKEN")
		if tokenString == "" {
			errorResponse.Meta.Status = false
			errorResponse.Meta.Message = "Missing authentication token"
			return ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse)
		}

		claims, err := o.authJwt.VerifyAccessToken(tokenString)
		if err != nil {
			errorResponse.Meta.Status = false

			switch {
			case errors.Is(err, auth.ErrTokenRevoked):
				errorResponse.Meta.Message = "Token has been revoked"
			case errors.Is(err, jwt.ErrTokenExpired):
				errorResponse.Meta.Message = "Token expired"
			case errors.Is(err, jwt.ErrTokenNotValidYet):
				errorResponse.Meta.Message = "Token not valid yet"
			case errors.Is(err, jwt.ErrTokenMalformed):
				errorResponse.Meta.Message = "Malformed token"
			case errors.Is(err, jwt.ErrTokenSignatureInvalid):
				errorResponse.Meta.Message = "Invalid token signature"
			default:
				errorResponse.Meta.Message = "Invalid token"
			}

			return ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse)
		}

		ctx.Locals("user", claims)
		return ctx.Next()
	}
}

func (o *Options) CheckOrigin(allowedOrigins string) fiber.Handler {
	allowedSet := make(map[string]struct{})
	for _, origin := range strings.Split(allowedOrigins, ",") {
		trimmed := strings.TrimRight(strings.TrimSpace(origin), "/")
		if trimmed != "" {
			allowedSet[trimmed] = struct{}{}
		}
	}

	return func(ctx *fiber.Ctx) error {
		switch ctx.Method() {
		case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
			origin := strings.TrimRight(strings.TrimSpace(ctx.Get("Origin")), "/")
			if origin != "" {
				if _, ok := allowedSet[origin]; !ok {
					var errorResponse response.ErrorResponseDefault
					errorResponse.Meta.Status = false
					errorResponse.Meta.Message = "Forbidden origin"
					return ctx.Status(fiber.StatusForbidden).JSON(errorResponse)
				}
			} else if strings.EqualFold(strings.TrimSpace(ctx.Get("Sec-Fetch-Site")), "cross-site") {
				var errorResponse response.ErrorResponseDefault
				errorResponse.Meta.Status = false
				errorResponse.Meta.Message = "Forbidden cross-site request"
				return ctx.Status(fiber.StatusForbidden).JSON(errorResponse)
			}
		}
		return ctx.Next()
	}
}
