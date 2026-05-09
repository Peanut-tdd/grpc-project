package main

import (
	"fmt"
	"github.com/pbuser/server/service/gateway"
	"log"
	"net"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	pb "github.com/pbuser/genproto/user"
	"github.com/pbuser/server/middleware"
	"github.com/pbuser/server/service"
	"google.golang.org/grpc"
)

const (
	Addr           = ":8080"
	NetWork        = "tcp"
	HttpAddr       = ":5002"
	DefaultTimeout = 5 * time.Second
)

func main() {
	listener, err := net.Listen(NetWork, Addr)
	if err != nil {
		log.Fatalf("net.Listen err: %v", err)
		return
	}

	defer listener.Close()
	defer middleware.CloseLogger()

	fmt.Println("server lister is ", listener.Addr())

	//拦截器，可注册日志，授权认证
	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			middleware.TimeoutStreamInterceptor(DefaultTimeout),
			grpc_auth.StreamServerInterceptor(middleware.AuthInterceptor),
			grpc_zap.StreamServerInterceptor(middleware.ZapInterceptor()),
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			middleware.TimeoutUnaryInterceptor(DefaultTimeout),
			grpc_auth.UnaryServerInterceptor(middleware.AuthInterceptor),
			grpc_zap.UnaryServerInterceptor(middleware.ZapInterceptor()),
		)),
	)

	pb.RegisterUserServiceServer(grpcServer, service.NewUserService())
	pb.RegisterStreamServiceServer(grpcServer, service.NewStreamService())
	pb.RegisterStreamClientServer(grpcServer, service.NewUploadService())
	pb.RegisterStreamServer(grpcServer, service.NewBothStreamServer())

	pb.RegisterGoodServer(grpcServer, service.NewGoodService())

	// 在 goroutine 中启动 gRPC 服务，防止阻塞
	go func() {
		fmt.Printf("gRPC server listening on %s\n", Addr)
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("grpcServer.Serve err: %v", err)
		}
	}()

	// 使用 gateway 把 grpcServer 转成 httpServer
	// 这里的 Addr 是 gRPC 的地址，"127.0.0.1:5002" 是 Gateway 的监听地址
	httpServer := gateway.ProvideHTTP("127.0.0.1"+Addr, "127.0.0.1"+HttpAddr, grpcServer)
	fmt.Printf("HTTP Gateway listening on %s\n", fmt.Sprintf("%s%s", "127.0.0.1", HttpAddr))
	if err = httpServer.ListenAndServe(); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
