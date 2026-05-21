package main

import (
	"context"
	"fmt"
	"github.com/pbuser/server/service/etcd"
	"github.com/pbuser/server/service/gateway"

	//"github.com/pbuser/server/service/gateway"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	pb "github.com/pbuser/genproto/user"
	"github.com/pbuser/server/middleware"
	"github.com/pbuser/server/service"
	servertrace "github.com/pbuser/server/trace"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

const (
	Addr           = ":8080"
	NetWork        = "tcp"
	HttpAddr       = ":5002"
	DefaultTimeout = 5 * time.Second
	SerName        = "grpc-demo"
)

var EndPoints = []string{"localhost:2379"}

func main() {

	go startPprof()

	ctx := context.Background()
	cleanup := servertrace.InitTracer(ctx)
	defer cleanup()

	//zaplogger:=middleware.ZapInterceptor()
	middleware.ZapInterceptor()

	ctx = servertrace.FuncCall(ctx, "main")
	listener, err := net.Listen(NetWork, Addr)
	if err != nil {
		log.Fatalf("net.Listen err: %v", err)
		return
	}

	defer listener.Close()
	defer middleware.CloseLogger()
	fmt.Println("server lister is ", listener.Addr())

	middleware.CtxInfof(ctx, "server lister is:%s", listener.Addr())

	//拦截器，可注册日志，授权认证
	grpcServer := grpc.NewServer(
		//metadata中获取client端的traceid
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			middleware.TimeoutStreamInterceptor(DefaultTimeout),
			grpc_auth.StreamServerInterceptor(middleware.AuthInterceptor),

			//记录全局日志
			//grpc_zap.StreamServerInterceptor(zaplogger),
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			middleware.TimeoutUnaryInterceptor(DefaultTimeout),
			grpc_auth.UnaryServerInterceptor(middleware.AuthInterceptor),
			//	grpc_zap.UnaryServerInterceptor(zaplogger),
		)),
	)

	pb.RegisterUserServiceServer(grpcServer, service.NewUserService())
	pb.RegisterStreamServiceServer(grpcServer, service.NewStreamService())
	pb.RegisterStreamClientServer(grpcServer, service.NewUploadService())
	pb.RegisterStreamServer(grpcServer, service.NewBothStreamServer())

	pb.RegisterGoodServer(grpcServer, service.NewGoodService())

	//etcd服务注册
	ser, err := etcd.NewServiceRegister(EndPoints, SerName, Addr, 60)
	if err != nil {
		log.Fatalf("etcd NewServiceRegister err: %v", err)
	}
	defer ser.Close()

	//基于反射的grcpurl
	reflection.Register(grpcServer)
	// 在 goroutine 中启动 gRPC 服务，防止阻塞
	go func() {
		fmt.Printf("gRPC server listening on %s\n", Addr)
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("grpcServer.Serve err: %v", err)
		}
	}()

	// 使用 gateway 把 grpcServer 转成 httpServer
	// 这里的 Addr 是 gRPC 的地址，"127.0.0.1:5002" 是 Gateway 的监听地址
	httpServer := gateway.NewGateway("127.0.0.1"+Addr, "127.0.0.1"+HttpAddr, grpcServer)
	fmt.Printf("HTTP Gateway listening on %s\n", fmt.Sprintf("%s%s", "127.0.0.1", HttpAddr))
	if err = httpServer.ListenAndServe(); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func startPprof() {
	mux := http.NewServeMux()
	mux.Handle("/debug/", http.DefaultServeMux)

	server := &http.Server{
		Addr:    "0.0.0.0:6060",
		Handler: mux,
	}

	log.Println("pprof server listening on :6060")
	if err := server.ListenAndServe(); err != nil {
		log.Printf("pprof server error: %v", err)
	}
}
