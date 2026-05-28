// Package token
package token

import (
	"maps"
	"time"

	"github.com/iconnor-code/cogo/cerrs"
	"github.com/iconnor-code/cogo/core"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const JwtTokenKey = "access_token"

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

func (j *JwtToken) GenerateToken(userInfo map[string]any) error {
	refreshToken := j.generateRefreshToken()
	accessToken, err := j.generateAccessToken(userInfo)
	if err != nil {
		return err
	}
	j.AccessToken = accessToken
	j.RefreshToken = refreshToken
	j.AccessExpireTime = time.Now().Add(time.Duration(j.config.Get("jwt.access_expire").(int)) * time.Hour)
	j.RefreshExpireTime = time.Now().Add(time.Duration(j.config.Get("jwt.refresh_expire").(int)) * time.Hour * 24)

	return nil
}

func (j *JwtToken) ParseToken(accessToken string) (map[string]any, error) {
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (any, error) {
		return []byte(j.config.Get("jwt.access_secret").(string)), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
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

func (j *JwtToken) generateAccessToken(userInfo map[string]any) (string, error) {
	t := jwt.New(jwt.SigningMethodHS256)

	claims := jwt.MapClaims{}
	maps.Copy(claims, userInfo)
	claims["exp"] = time.Now().Add(time.Duration(j.config.Get("jwt.access_expire").(int)) * time.Hour).Unix()

	t.Claims = claims
	s, err := t.SignedString([]byte(j.config.Get("jwt.access_secret").(string)))
	return s, err
}

func (j *JwtToken) generateRefreshToken() string {
	return uuid.New().String()
}
