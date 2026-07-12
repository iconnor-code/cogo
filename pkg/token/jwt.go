// Package token
package token

import (
	"encoding/json"
	"fmt"
	"maps"
	"math"
	"strings"
	"time"

	"github.com/iconnor-code/cogo/cerrs"
	"github.com/iconnor-code/cogo/core"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	AuthorizationHeader = "authorization"
	BearerScheme        = "Bearer"
)

const (
	ClaimExpiresAt = "exp"
	ClaimID        = "jti"
)

type JwtToken struct {
	config            Config
	AccessToken       string
	AccessTokenID     string
	RefreshToken      string
	AccessExpireTime  time.Time
	RefreshExpireTime time.Time
}

type Config interface {
	GetJWT() core.JWTConfig
}

func NewJwtToken(config Config) *JwtToken {
	return &JwtToken{config: config}
}

// ExtractBearerToken returns the JWT carried by a standard Authorization header.
func ExtractBearerToken(value string) (string, error) {
	parts := strings.Fields(value)
	if len(parts) != 2 || !strings.EqualFold(parts[0], BearerScheme) || parts[1] == "" {
		return "", cerrs.New("authorization header must use Bearer scheme")
	}
	return parts[1], nil
}

func (j *JwtToken) GenerateToken(userInfo map[string]any) error {
	refreshToken := j.generateRefreshToken()
	accessToken, accessClaims, err := j.generateAccessToken(userInfo)
	if err != nil {
		return err
	}
	jwtConf := j.config.GetJWT()
	accessExpire := jwtConf.AccessExpire
	refreshExpire := jwtConf.RefreshExpire
	j.AccessToken = accessToken
	j.AccessTokenID, _ = ClaimsString(accessClaims, ClaimID)
	j.RefreshToken = refreshToken
	j.AccessExpireTime = time.Now().Add(time.Duration(accessExpire) * time.Hour)
	j.RefreshExpireTime = time.Now().Add(time.Duration(refreshExpire) * time.Hour * 24)

	return nil
}

func (j *JwtToken) ParseToken(accessToken string) (map[string]any, error) {
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (any, error) {
		return []byte(j.config.GetJWT().AccessSecret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithExpirationRequired())
	if err != nil {
		return nil, cerrs.Wrap(err)
	}
	if token == nil || !token.Valid {
		return nil, cerrs.New("invalid access token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, cerrs.New("invalid access token claims")
	}
	return claims, nil
}

func (j *JwtToken) generateAccessToken(userInfo map[string]any) (string, jwt.MapClaims, error) {
	jwtConf := j.config.GetJWT()
	accessExpire := jwtConf.AccessExpire
	secret := jwtConf.AccessSecret
	t := jwt.New(jwt.SigningMethodHS256)

	claims := jwt.MapClaims{}
	maps.Copy(claims, userInfo)
	claims[ClaimExpiresAt] = time.Now().Add(time.Duration(accessExpire) * time.Hour).Unix()
	claims[ClaimID] = uuid.NewString()

	t.Claims = claims
	s, err := t.SignedString([]byte(secret))
	return s, claims, err
}

func (j *JwtToken) generateRefreshToken() string {
	return uuid.New().String()
}

func ClaimsString(claims map[string]any, key string) (string, error) {
	value, ok := claims[key]
	if !ok {
		return "", fmt.Errorf("%s is required", key)
	}
	s, ok := value.(string)
	if !ok || s == "" {
		return "", fmt.Errorf("%s is invalid", key)
	}
	return s, nil
}

func ClaimsExpiresAt(claims map[string]any) (time.Time, error) {
	value, ok := claims[ClaimExpiresAt]
	if !ok {
		return time.Time{}, fmt.Errorf("%s is required", ClaimExpiresAt)
	}

	var unix float64
	switch v := value.(type) {
	case float64:
		unix = v
	case float32:
		unix = float64(v)
	case int:
		unix = float64(v)
	case int64:
		unix = float64(v)
	case uint64:
		unix = float64(v)
	case json.Number:
		parsed, err := v.Float64()
		if err != nil {
			return time.Time{}, fmt.Errorf("%s is invalid: %w", ClaimExpiresAt, err)
		}
		unix = parsed
	default:
		return time.Time{}, fmt.Errorf("%s has unsupported type %T", ClaimExpiresAt, value)
	}
	if unix <= 0 || math.Trunc(unix) != unix {
		return time.Time{}, fmt.Errorf("%s is invalid", ClaimExpiresAt)
	}
	return time.Unix(int64(unix), 0), nil
}
