package auth

import (
	"context"
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

type Token struct {
	Value   string
	TraceId string
}

const headerAuthorize string = "authorization"
const traceId string = "traceId"

// GetRequestMetadata 获取当前请求认证所需的元数据
func (t *Token) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{headerAuthorize: t.Value, traceId: t.TraceId}, nil
}

// RequireTransportSecurity 是否需要基于 TLS 认证进行安全传输
func (t *Token) RequireTransportSecurity() bool {
	return false
}

func CreateAuth() (*TokenPair, error) {

	jwtManager = NewJwtManager()

	tokens, err := jwtManager.GenerateTokens(123, "admin")

	return tokens, err

}
