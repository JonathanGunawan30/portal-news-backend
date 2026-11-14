package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/JonathanGunawan30/portal-news-backend.git/config"
	"github.com/JonathanGunawan30/portal-news-backend.git/internal/core/domain/entity"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrTokenRevoked            = errors.New("token has been revoked")
	ErrSigningKeyNotConfigured = errors.New("jwt signing key is not configured")
)

type Jwt interface {
	GenerateToken(data *entity.JwtData) (string, int64, error)
	VerifyAccessToken(token string) (*entity.JwtData, error)
	RevokeToken(token string) error
	InvalidateUserTokens(userID int64)
}

type Options struct {
	signingKey        string
	issuer            string
	mu                sync.RWMutex
	revokedTokens     map[string]time.Time
	userMinIssuedNano map[int64]int64
}

func NewJwt(cfg *config.Config) Jwt {
	return &Options{
		signingKey:        cfg.App.JwtSecretKey,
		issuer:            cfg.App.JwtIssuer,
		revokedTokens:     make(map[string]time.Time),
		userMinIssuedNano: make(map[int64]int64),
	}
}

func (o *Options) GenerateToken(data *entity.JwtData) (string, int64, error) {
	if strings.TrimSpace(o.signingKey) == "" {
		return "", 0, ErrSigningKeyNotConfigured
	}
	if data == nil || data.UserID <= 0 {
		return "", 0, errors.New("invalid jwt user data")
	}

	now := time.Now().UTC()
	expiredAt := now.Add(24 * time.Hour)

	jti := data.RegisteredClaims.ID
	if jti == "" {
		jti = uuid.NewString()
	}

	data.IssuedAtNano = now.UnixNano()
	data.RegisteredClaims = jwt.RegisteredClaims{
		ID:        jti,
		ExpiresAt: jwt.NewNumericDate(expiredAt),
		Issuer:    o.issuer,
		NotBefore: jwt.NewNumericDate(now),
		IssuedAt:  jwt.NewNumericDate(now),
		Subject:   strconv.FormatInt(data.UserID, 10),
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, data)
	signed, err := tok.SignedString([]byte(o.signingKey))
	if err != nil {
		return "", 0, err
	}
	return signed, expiredAt.Unix(), nil
}

func (o *Options) VerifyAccessToken(tokenStr string) (*entity.JwtData, error) {
	if strings.TrimSpace(o.signingKey) == "" {
		return nil, ErrSigningKeyNotConfigured
	}

	opts := []jwt.ParserOption{
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	}
	if o.issuer != "" {
		opts = append(opts, jwt.WithIssuer(o.issuer))
	}

	parsed, err := jwt.ParseWithClaims(tokenStr, &entity.JwtData{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(o.signingKey), nil
	}, opts...)
	if err != nil {
		return nil, err
	}

	claims, ok := parsed.Claims.(*entity.JwtData)
	if !ok || !parsed.Valid || claims.UserID <= 0 {
		return nil, fmt.Errorf("invalid token claims")
	}

	if o.isRevoked(claims, tokenStr) {
		return nil, ErrTokenRevoked
	}

	return claims, nil
}

func (o *Options) RevokeToken(tokenStr string) error {
	if strings.TrimSpace(tokenStr) == "" || strings.TrimSpace(o.signingKey) == "" {
		return nil
	}

	opts := []jwt.ParserOption{
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	}
	if o.issuer != "" {
		opts = append(opts, jwt.WithIssuer(o.issuer))
	}

	parsed, err := jwt.ParseWithClaims(tokenStr, &entity.JwtData{}, func(t *jwt.Token) (any, error) {
		return []byte(o.signingKey), nil
	}, opts...)
	if err != nil {
		return err
	}

	claims, ok := parsed.Claims.(*entity.JwtData)
	if !ok || !parsed.Valid {
		return nil
	}

	exp := time.Now().UTC().Add(24 * time.Hour)
	if claims.ExpiresAt != nil {
		exp = claims.ExpiresAt.Time.UTC()
	}
	now := time.Now().UTC()
	if !exp.After(now) {
		return nil
	}

	key := revocationKey(claims.ID, tokenStr)
	o.mu.Lock()
	defer o.mu.Unlock()
	for k, v := range o.revokedTokens {
		if !v.After(now) {
			delete(o.revokedTokens, k)
		}
	}
	o.revokedTokens[key] = exp
	return nil
}

func (o *Options) InvalidateUserTokens(userID int64) {
	if userID <= 0 {
		return
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	o.userMinIssuedNano[userID] = time.Now().UTC().UnixNano()
}

func (o *Options) isRevoked(claims *entity.JwtData, tokenStr string) bool {
	now := time.Now().UTC()
	key := revocationKey(claims.ID, tokenStr)

	o.mu.RLock()
	defer o.mu.RUnlock()

	if exp, exists := o.revokedTokens[key]; exists && exp.After(now) {
		return true
	}
	if minNano, exists := o.userMinIssuedNano[claims.UserID]; exists {
		if claims.IssuedAtNano == 0 || claims.IssuedAtNano <= minNano {
			return true
		}
	}
	return false
}

func revocationKey(jti, tokenStr string) string {
	if jti != "" {
		return "jti:" + jti
	}
	sum := sha256.Sum256([]byte(tokenStr))
	return "tok:" + hex.EncodeToString(sum[:])
}
