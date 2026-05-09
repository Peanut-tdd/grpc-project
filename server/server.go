package main

import (
	"context"
	"fmt"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/pbuser/server/middleware"
	"github.com/pbuser/server/service"
	"log"
	"net"
	"time"

	pb "github.com/pbuser/genproto/user"
	"google.golang.org/grpc"
)

const (
	Addr    = ":8080"
	NetWork = "tcp"
)

func main() {
	lister, err := net.Listen(NetWork, Addr)
	if err != nil {
		log.Fatalf("net.Listen err: %v", err)
		return
	}

	defer lister.Close()

	fmt.Println("server lister is ", lister.Addr())

	grpcServer := grpc.NewServer(grpc.StreamInterceptor(
		grpc_middleware.ChainStreamServer(grpc_auth.StreamServerInterceptor(middleware.AuthInterceptor))),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(grpc_auth.UnaryServerInterceptor(middleware.AuthInterceptor))))

	pb.RegisterUserServiceServer(grpcServer, service.NewUserService())
	pb.RegisterStreamServiceServer(grpcServer, service.NewStreamService())
	pb.RegisterStreamClientServer(grpcServer, service.NewUploadService())
	pb.RegisterStreamServer(grpcServer, service.NewBothStreamServer())

	pb.RegisterGoodServer(grpcServer, service.NewGoodService())

	err = grpcServer.Serve(lister)
	if err != nil {
		log.Fatalf("grpcServer.Serve err: %v", err)
	}
}

func TimeoutStreamInterceptor(timeout time.Duration) grpc.StreamServerInterceptor {

	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		ctx := ss.Context()

		// 如果客户端已经设置了 deadline，则不再覆盖
		if _, ok := ctx.Deadline(); ok {
			return handler(srv, ss)
		}

		// 设置服务端默认超时
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		// 包装 ServerStream，注入新的 context
		wrapped := &wrappedServerStream{
			ServerStream: ss,
			ctx:          ctx,
		}

		return handler(srv, wrapped)
	}
}

// 包装 ServerStream
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

func TimeoutUnaryInterceptor(timeout time.Duration) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {

		// 如果客户端已经设置了 deadline，就不覆盖
		if _, ok := ctx.Deadline(); ok {
			return handler(ctx, req)
		}

		// 设置服务端默认超时
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		return handler(ctx, req)
	}
}
