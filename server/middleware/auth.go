package middleware

import (
	"context"
	"errors"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/pbuser/server/jwt"
)

var UserKey string

type TokenInfo struct {
	UserId uint `json:"user_id"`
}

func AuthInterceptor(ctx context.Context) (context.Context, error) {
	token, err := grpc_auth.AuthFromMD(ctx, "bearer")
	if err != nil {
		return nil, err
	}

	//从metadata中获取客户端设置的字段
	//md, ok := metadata.FromIncomingContext(ctx)
	//if !ok {
	//	return nil, err
	//}
	//
	//b, _ := json.Marshal(md)
	//fmt.Println(string(b))
	//
	//
	//var mataToken string
	//var traceId string
	//if value, ok := md["authorization"]; ok {
	//	mataToken = value[0]
	//}
	//if value, ok := md["traceid"]; ok {
	//	traceId = value[0]
	//}
	//
	//fmt.Printf("mataToken: %s,traceId:%s\n", mataToken, traceId)

	//fmt.Printf("token is %v\n", token)

	jwtManager := jwt.NewJwtManager()

	claims, err := jwtManager.ParseToken(token)
	if err != nil {
		return nil, err
	}

	//fmt.Println(claims.UserId)
	return context.WithValue(ctx, UserKey, &TokenInfo{UserId: claims.UserId}), nil
}

func GetUserInfo(ctx context.Context) (*TokenInfo, error) {

	user, ok := ctx.Value(UserKey).(*TokenInfo)
	if !ok {
		return nil, errors.New("get user info fail")
	}

	return user, nil
}
