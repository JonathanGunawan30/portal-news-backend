package auth

import (
	"testing"
	"time"

	"github.com/JonathanGunawan30/portal-news-backend.git/config"
	"github.com/JonathanGunawan30/portal-news-backend.git/internal/core/domain/entity"
)

func testConfig(secret, issuer string) *config.Config {
	return &config.Config{
		App: config.App{
			JwtSecretKey: secret,
			JwtIssuer:    issuer,
		},
	}
}

func TestJwtGenerateVerifyAndRevoke(t *testing.T) {
	j := NewJwt(testConfig("super-secret-key-at-least-32-chars!!", "portal-news"))

	tok, _, err := j.GenerateToken(&entity.JwtData{UserID: 42})
	if err != nil {
		t.Fatalf("unexpected generate error: %v", err)
	}

	claims, err := j.VerifyAccessToken(tok)
	if err != nil {
		t.Fatalf("unexpected verify error: %v", err)
	}
	if claims.UserID != 42 {
		t.Fatalf("expected UserID 42, got %d", claims.UserID)
	}

	// Wrong issuer must be rejected
	otherIssuerJwt := NewJwt(testConfig("super-secret-key-at-least-32-chars!!", "other-issuer"))
	if _, err := otherIssuerJwt.VerifyAccessToken(tok); err == nil {
		t.Fatalf("expected token from different issuer to be rejected")
	}

	// Revoke token on logout
	if err := j.RevokeToken(tok); err != nil {
		t.Fatalf("unexpected revoke error: %v", err)
	}
	if _, err := j.VerifyAccessToken(tok); err == nil {
		t.Fatalf("expected revoked token to be rejected")
	}
}

func TestJwtInvalidateUserTokens(t *testing.T) {
	j := NewJwt(testConfig("super-secret-key-at-least-32-chars!!", "portal-news"))

	oldTok, _, err := j.GenerateToken(&entity.JwtData{UserID: 7})
	if err != nil {
		t.Fatalf("unexpected generate error: %v", err)
	}

	time.Sleep(2 * time.Millisecond)
	j.InvalidateUserTokens(7)
	time.Sleep(2 * time.Millisecond)

	if _, err := j.VerifyAccessToken(oldTok); err == nil {
		t.Fatalf("expected token issued before InvalidateUserTokens to be rejected")
	}

	newTok, _, err := j.GenerateToken(&entity.JwtData{UserID: 7})
	if err != nil {
		t.Fatalf("unexpected generate error for new token: %v", err)
	}
	if _, err := j.VerifyAccessToken(newTok); err != nil {
		t.Fatalf("expected new token issued after InvalidateUserTokens to be valid, got: %v", err)
	}
}

func TestJwtEmptySigningKeyRejected(t *testing.T) {
	j := NewJwt(testConfig("", "portal-news"))
	if _, _, err := j.GenerateToken(&entity.JwtData{UserID: 1}); err == nil {
		t.Fatalf("expected error with empty signing key")
	}
	if _, err := j.VerifyAccessToken("a.b.c"); err == nil {
		t.Fatalf("expected error verifying with empty signing key")
	}
}
