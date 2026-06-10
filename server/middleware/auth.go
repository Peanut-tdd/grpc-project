package middleware

import (
	"context"
	"errors"
	"fmt"
	"strings"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/pbuser/server/jwt"
	"google.golang.org/grpc"
)

var UserKey string

type TokenInfo struct {
	UserId uint `json:"user_id"`
}

// AuthServices 需要认证的服务名集合
var AuthServices = map[string]bool{
	"UserService":  true,
	"StreamService": true,
	"StreamClient": true,
	"Stream":       true,
}

// extractService 从 FullMethod 提取服务名
// "/user.UserService/CreateUser" → "UserService"
func extractService(fullMethod string) string {
	// 去掉前缀 "/"
	method := strings.TrimPrefix(fullMethod, "/")
	// 按 "/" 分割，第一部分是 "package.ServiceName"
	parts := strings.SplitN(method, "/", 2)
	// 再按 "." 分割，取最后一段即服务名
	dotParts := strings.SplitN(parts[0], ".", 2)
	if len(dotParts) == 2 {
		return dotParts[1]
	}
	return ""
}

// AuthUnaryInterceptor 一元 RPC 认证拦截器，仅对 AuthServices 中的服务校验 JWT
func AuthUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		svcName := extractService(info.FullMethod)
		fmt.Printf("AuthUnaryInterceptor: method=%s, service=%s\n", info.FullMethod, svcName)

		if !AuthServices[svcName] {
			// 不需要认证的服务，直接放行
			return handler(ctx, req)
		}

		token, err := grpc_auth.AuthFromMD(ctx, "bearer")
		if err != nil {
			return nil, err
		}

		jwtManager := jwt.NewJwtManager()
		claims, err := jwtManager.ParseToken(token)
		if err != nil {
			return nil, err
		}

		ctx = context.WithValue(ctx, UserKey, &TokenInfo{UserId: claims.UserId})
		return handler(ctx, req)
	}
}

// AuthStreamInterceptor 流式 RPC 认证拦截器，仅对 AuthServices 中的服务校验 JWT
func AuthStreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		svcName := extractService(info.FullMethod)
		fmt.Printf("AuthStreamInterceptor: method=%s, service=%s\n", info.FullMethod, svcName)

		if !AuthServices[svcName] {
			// 不需要认证的服务，直接放行
			return handler(srv, ss)
		}

		ctx := ss.Context()
		token, err := grpc_auth.AuthFromMD(ctx, "bearer")
		if err != nil {
			return err
		}

		jwtManager := jwt.NewJwtManager()
		claims, err := jwtManager.ParseToken(token)
		if err != nil {
			return err
		}

		ctx = context.WithValue(ctx, UserKey, &TokenInfo{UserId: claims.UserId})
		wrapped := &wrappedAuthStream{ServerStream: ss, ctx: ctx}
		return handler(srv, wrapped)
	}
}

type wrappedAuthStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedAuthStream) Context() context.Context {
	return w.ctx
}

func GetUserInfo(ctx context.Context) (*TokenInfo, error) {
	user, ok := ctx.Value(UserKey).(*TokenInfo)
	if !ok {
		return nil, errors.New("get user info fail")
	}
	return user, nil
}
