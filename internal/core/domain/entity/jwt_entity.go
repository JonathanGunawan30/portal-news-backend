package entity

import "github.com/golang-jwt/jwt/v5"

type JwtData struct {
	UserID       int64 `json:"user_id"`
	IssuedAtNano int64 `json:"iat_ns,omitempty"`
	jwt.RegisteredClaims
}
