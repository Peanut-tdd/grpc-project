package middleware

import (
	"context"
	"fmt"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/pbuser/server/jwt"
)

var userKey string

type TokenInfo struct {
	UserId uint `json:"user_id"`
}

func AuthInterceptor(ctx context.Context) (context.Context, error) {
	token, err := grpc_auth.AuthFromMD(ctx, "bearer")
	if err != nil {
		return nil, err
	}

	fmt.Printf("token is %v\n", token)

	jwtManager := jwt.NewJwtManager()

	claims, err := jwtManager.ParseToken(token)
	if err != nil {
		return nil, err
	}

	fmt.Println(claims.UserId)
	return context.WithValue(ctx, userKey, TokenInfo{UserId: claims.UserId}), nil
}
