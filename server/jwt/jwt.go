package jwt

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/cast"
	"sync"
	"time"
)

type JwtManager struct {
	Secret        string        `json:"secret"`
	AccessExpire  time.Duration `json:"access_expire"`
	RefreshExpire time.Duration `json:"refresh_expire"`
}

type UserClaims struct {
	UserId   uint   `json:"user_id"`
	Username string `json:"user_name"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

var jwtManager *JwtManager
var synconce sync.Once

func NewJwtManager() *JwtManager {
	synconce.Do(func() {
		jwtManager = &JwtManager{
			Secret:        "123456",
			AccessExpire:  time.Duration(7) * time.Minute,
			RefreshExpire: time.Duration(14) * time.Hour,
		}
	})
	return jwtManager
}

func (jm *JwtManager) GenerateTokens(userId uint, username string) (*TokenPair, error) {
	accessToken, err := jm.generateAccessToken(userId, username)

	if err != nil {
		return nil, err
	}

	refreshToken, err := jm.RefreshTokens(userId)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    cast.ToInt(jm.AccessExpire.Seconds()),
	}, nil

}

func (jm *JwtManager) generateAccessToken(userId uint, username string) (string, error) {

	claims := &UserClaims{
		UserId:   userId,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(jm.AccessExpire)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jm.Secret))

}

func (jm *JwtManager) RefreshTokens(userId uint) (string, error) {
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(jm.RefreshExpire)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ID:        cast.ToString(userId),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jm.Secret))
}

func (jm *JwtManager) ParseToken(tokenString string) (*UserClaims, error) {

	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {

		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("无效的签名方法")
		}
		return []byte(jm.Secret), nil
	})

	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {

		return claims, nil
	}

	return nil, errors.New("无效的令牌")
}
