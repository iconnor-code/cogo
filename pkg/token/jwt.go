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
	accessExpire, err := core.GetInt(j.config, "jwt.access_expire")
	if err != nil {
		return cerrs.Wrap(err)
	}
	refreshExpire, err := core.GetInt(j.config, "jwt.refresh_expire")
	if err != nil {
		return cerrs.Wrap(err)
	}
	j.AccessToken = accessToken
	j.RefreshToken = refreshToken
	j.AccessExpireTime = time.Now().Add(time.Duration(accessExpire) * time.Hour)
	j.RefreshExpireTime = time.Now().Add(time.Duration(refreshExpire) * time.Hour * 24)

	return nil
}

func (j *JwtToken) ParseToken(accessToken string) (map[string]any, error) {
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (any, error) {
		secret, err := core.GetString(j.config, "jwt.access_secret")
		if err != nil {
			return nil, cerrs.Wrap(err)
		}
		return []byte(secret), nil
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
	accessExpire, err := core.GetInt(j.config, "jwt.access_expire")
	if err != nil {
		return "", cerrs.Wrap(err)
	}
	secret, err := core.GetString(j.config, "jwt.access_secret")
	if err != nil {
		return "", cerrs.Wrap(err)
	}
	t := jwt.New(jwt.SigningMethodHS256)

	claims := jwt.MapClaims{}
	maps.Copy(claims, userInfo)
	claims["exp"] = time.Now().Add(time.Duration(accessExpire) * time.Hour).Unix()

	t.Claims = claims
	s, err := t.SignedString([]byte(secret))
	return s, err
}

func (j *JwtToken) generateRefreshToken() string {
	return uuid.New().String()
}
