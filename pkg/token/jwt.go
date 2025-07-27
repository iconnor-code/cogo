// Package token
package token

import (
	"math"
	"time"

	"github.com/iconnor-code/cogo/cerrs"
	"github.com/iconnor-code/cogo/core"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const JwtTokenKey = "access_token"

type User struct {
	ID uint32
}

type JwtToken struct {
	config            core.IConfig
	AccessToken       string
	RefreshToken      string
	AccessExpireTime  time.Time
	RefreshExpireTime time.Time
}

func NewJwtToken(config core.IConfig) *JwtToken {
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
	j.AccessExpireTime = time.Now().Add(time.Duration(j.config.Get("jwt.access_expire").(int)) * time.Hour)
	j.RefreshExpireTime = time.Now().Add(time.Duration(j.config.Get("jwt.refresh_expire").(int)) * time.Hour * 24)

	return nil
}

func (j *JwtToken) ParseToken(accessToken string) (*User, error) {
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (any, error) {
		return []byte(j.config.Get("jwt.access_secret").(string)), nil
	})
	if err != nil {
		return nil, cerrs.Wrap(err)
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, cerrs.New("invalid access token claims")
	}

	userID, ok := claims["user_id"].(float64)
	if !ok {
		return nil, cerrs.New("invalid user_id in token claims")
	}
	if userID < 0 || userID > math.MaxUint32 {
		return nil, cerrs.New("user_id out of range")
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
	claims["exp"] = time.Now().Add(time.Duration(j.config.Get("jwt.access_expire").(int)) * time.Hour).Unix()
	t.Claims = claims
	s, err := t.SignedString([]byte(j.config.Get("jwt.access_secret").(string)))
	return s, err
}

func (j *JwtToken) generateRefreshToken() string {
	return uuid.New().String()
}
