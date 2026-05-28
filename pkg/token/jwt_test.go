package token

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type testConfig struct {
	data map[string]any
}

func (c *testConfig) Get(key string) any { return c.data[key] }
func (c *testConfig) ReLoad() error      { return nil }

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
	j := NewJwtToken(&testConfig{data: map[string]any{
		"jwt.access_secret": "secret",
	}})
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

func TestParseTokenRejectsExpiredToken(t *testing.T) {
	j := NewJwtToken(&testConfig{data: map[string]any{
		"jwt.access_secret": "secret",
	}})
	accessToken := signToken(t, jwt.SigningMethodHS256, "secret", jwt.MapClaims{
		"exp": time.Now().Add(-time.Hour).Unix(),
	})

	if _, err := j.ParseToken(accessToken); err == nil {
		t.Fatalf("expected expired token error")
	}
}

func TestParseTokenRejectsUnexpectedSigningMethod(t *testing.T) {
	j := NewJwtToken(&testConfig{data: map[string]any{
		"jwt.access_secret": "secret",
	}})
	accessToken := signToken(t, jwt.SigningMethodHS384, "secret", jwt.MapClaims{
		"exp": time.Now().Add(time.Hour).Unix(),
	})

	if _, err := j.ParseToken(accessToken); err == nil {
		t.Fatalf("expected signing method error")
	}
}
