package token

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/iconnor-code/cogo/core"
	cogoconfig "github.com/iconnor-code/cogo/core/impl/config"
)

func signToken(t *testing.T, method jwt.SigningMethod, secret string, claims jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(method, claims)
	s, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return s
}

func TestParseTokenSuccess(t *testing.T) {
	j := NewJwtToken(&cogoconfig.Config{Config: core.Config{JWT: core.JWTConfig{AccessSecret: "secret"}}})
	accessToken := signToken(t, jwt.SigningMethodHS256, "secret", jwt.MapClaims{
		"user_id":    float64(123),
		"user_email": "u@test.com",
		"exp":        time.Now().Add(time.Hour).Unix(),
	})

	claims, err := j.ParseToken(accessToken)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if claims["user_email"] != "u@test.com" {
		t.Fatalf("unexpected user_email: %v", claims["user_email"])
	}
}

func TestGenerateTokenAddsTokenIDAndExpiration(t *testing.T) {
	j := NewJwtToken(&cogoconfig.Config{Config: core.Config{JWT: core.JWTConfig{
		AccessSecret:  "secret",
		AccessExpire:  1,
		RefreshExpire: 1,
	}}})

	if err := j.GenerateToken(map[string]any{"user_id": uint32(123)}); err != nil {
		t.Fatalf("generate token: %v", err)
	}
	if j.AccessTokenID == "" {
		t.Fatalf("expected access token id")
	}

	claims, err := j.ParseToken(j.AccessToken)
	if err != nil {
		t.Fatalf("parse generated token: %v", err)
	}
	tokenID, err := ClaimsString(claims, ClaimID)
	if err != nil {
		t.Fatalf("get token id: %v", err)
	}
	if tokenID != j.AccessTokenID {
		t.Fatalf("token id mismatch: %s != %s", tokenID, j.AccessTokenID)
	}
	if _, err := ClaimsExpiresAt(claims); err != nil {
		t.Fatalf("get expiration: %v", err)
	}
}

func TestParseTokenRejectsExpiredToken(t *testing.T) {
	j := NewJwtToken(&cogoconfig.Config{Config: core.Config{JWT: core.JWTConfig{AccessSecret: "secret"}}})
	accessToken := signToken(t, jwt.SigningMethodHS256, "secret", jwt.MapClaims{
		"exp": time.Now().Add(-time.Hour).Unix(),
	})

	if _, err := j.ParseToken(accessToken); err == nil {
		t.Fatalf("expected expired token error")
	}
}

func TestParseTokenRejectsMissingExpiration(t *testing.T) {
	j := NewJwtToken(&cogoconfig.Config{Config: core.Config{JWT: core.JWTConfig{AccessSecret: "secret"}}})
	accessToken := signToken(t, jwt.SigningMethodHS256, "secret", jwt.MapClaims{
		"user_id": float64(123),
	})

	if _, err := j.ParseToken(accessToken); err == nil {
		t.Fatalf("expected missing expiration error")
	}
}

func TestParseTokenRejectsUnexpectedSigningMethod(t *testing.T) {
	j := NewJwtToken(&cogoconfig.Config{Config: core.Config{JWT: core.JWTConfig{AccessSecret: "secret"}}})
	accessToken := signToken(t, jwt.SigningMethodHS384, "secret", jwt.MapClaims{
		"exp": time.Now().Add(time.Hour).Unix(),
	})

	if _, err := j.ParseToken(accessToken); err == nil {
		t.Fatalf("expected signing method error")
	}
}
