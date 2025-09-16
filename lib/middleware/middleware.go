package middleware

import (
	"errors"
	"strings"

	"github.com/JonathanGunawan30/portal-news-backend.git/config"
	"github.com/JonathanGunawan30/portal-news-backend.git/internal/adapter/handler/response"
	"github.com/JonathanGunawan30/portal-news-backend.git/lib/auth"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type Middleware interface {
	CheckToken() fiber.Handler
}

type Options struct {
	authJwt auth.Jwt
}

func NewMiddleware(cfg *config.Config) Middleware {
	return &Options{authJwt: auth.NewJwt(cfg)}
}

func (o *Options) CheckToken() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var errorResponse response.ErrorResponseDefault

		authHeader := strings.TrimSpace(ctx.Get("Authorization"))

		if authHeader == "" {
			errorResponse.Meta.Status = false
			errorResponse.Meta.Message = "Missing Authorization Header"
			return ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse)
		}

		parts := strings.Fields(authHeader)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			errorResponse.Meta.Status = false
			errorResponse.Meta.Message = "Invalid Authorization header format"
			return ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse)
		}

		tokenString := parts[1]
		claims, err := o.authJwt.VerifyAccessToken(tokenString)

		if err != nil {
			errorResponse.Meta.Status = false

			switch {
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
