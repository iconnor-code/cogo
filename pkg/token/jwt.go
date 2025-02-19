package token

import (
	"math"
	"time"

	"github.com/iconnor-code/cogo/pkg/config"
	"github.com/iconnor-code/cogo/pkg/cerr"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

const JWT_TOKEN_KEY = "access_token"

type User struct {
	ID uint32
}

type JwtToken struct {
	config            *config.JwtTokenConfig
	AccessToken       string
	RefreshToken      string
	AccessExpireTime  time.Time
	RefreshExpireTime time.Time
}

func NewJwtToken(config *config.JwtTokenConfig) *JwtToken {
	return &JwtToken{config: config}
}

func (j *JwtToken) GenerateToken(user *User) error {
	refreshToken := j.generateRefreshToken()
	accessToken, err := j.generateAccessToken(user)
	if err != nil {
		return err
	}
	j.AccessToken = accessToken
	j.RefreshToken = refreshToken
	j.AccessExpireTime = time.Now().Add(time.Duration(j.config.AccessExpire) * time.Hour)
	j.RefreshExpireTime = time.Now().Add(time.Duration(j.config.RefreshExpire) * time.Hour * 24)

	return nil
}

func (j *JwtToken) ParseToken(accessToken string) (*User, error) {
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.config.AccessSecret), nil
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, cerr.NewClientError("invalid access token claims", nil)
	}

	userID, ok := claims["user_id"].(float64)
	if !ok {
		return nil, cerr.NewClientError("invalid user_id in token claims", nil)
	}
	if userID < 0 || userID > math.MaxUint32 {
		return nil, cerr.NewClientError("user_id out of range", nil)
	}
	return &User{
		ID: uint32(userID),
	}, nil
}

func (j *JwtToken) generateAccessToken(user *User) (string, error) {
	t := jwt.New(jwt.SigningMethodHS256)
	claims := jwt.MapClaims{
		"user_id": user.ID,
	}
	claims["exp"] = time.Now().Add(time.Duration(j.config.AccessExpire) * time.Hour).Unix()
	t.Claims = claims
	s, err := t.SignedString([]byte(j.config.AccessSecret))
	return s, err
}

func (j *JwtToken) generateRefreshToken() string {
	return uuid.New().String()
}
