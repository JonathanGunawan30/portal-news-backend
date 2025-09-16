package auth

import (
	"fmt"
	"time"

	"github.com/JonathanGunawan30/portal-news-backend.git/config"
	"github.com/JonathanGunawan30/portal-news-backend.git/internal/core/domain/entity"
	"github.com/golang-jwt/jwt/v5"
)

type Jwt interface {
	GenerateToken(data *entity.JwtData) (string, int64, error)
	VerifyAccessToken(token string) (*entity.JwtData, error)
}

type Options struct {
	signingKey string
	issuer     string
}

func NewJwt(cfg *config.Config) Jwt {
	return &Options{
		signingKey: cfg.App.JwtSecretKey,
		issuer:     cfg.App.JwtIssuer,
	}
}

func (o *Options) GenerateToken(data *entity.JwtData) (string, int64, error) {
	now := time.Now().UTC()
	expiredAt := now.Add(24 * time.Hour)

	data.RegisteredClaims = jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(expiredAt),
		Issuer:    o.issuer,
		NotBefore: jwt.NewNumericDate(now),
		IssuedAt:  jwt.NewNumericDate(now),
		Subject:   fmt.Sprint(data.UserID),
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, data)
	signed, err := tok.SignedString([]byte(o.signingKey))
	if err != nil {
		return "", 0, err
	}
	return signed, expiredAt.Unix(), nil
}

func (o *Options) VerifyAccessToken(tokenStr string) (*entity.JwtData, error) {
	parsed, err := jwt.ParseWithClaims(tokenStr, &entity.JwtData{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(o.signingKey), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := parsed.Claims.(*entity.JwtData)
	if !ok || !parsed.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}
	return claims, nil
}
