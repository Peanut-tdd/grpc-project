package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pbuser/client/auth"
	"github.com/pbuser/client/etcd"
	"github.com/pbuser/client/router"
	clienttrace "github.com/pbuser/client/trace"
	"github.com/pbuser/client/zap"
	pb "github.com/pbuser/genproto/user"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/resolver"
	"log"
	"net/http"
)

var (
	uGrpcClient    pb.UserServiceClient
	sGrpcClient    pb.StreamServiceClient
	cGrpcClient    pb.StreamClientClient
	bothGrpcClient pb.StreamClient
	gGrpcClient    pb.GoodClient
)

var EtcdEndpoints = []string{"localhost:2379"}

func unaryInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	var p peer.Peer
	if err := invoker(ctx, method, req, reply, cc, append(opts, grpc.Peer(&p))...); err != nil {
		return err
	}
	if p.Addr != nil {
		fmt.Printf("Unary call: %s -> Server Address: %s\n", method, p.Addr.String())
	}
	return nil
}

type wrappedStream struct {
	grpc.ClientStream
	method  string
	printed bool
}

func (w *wrappedStream) RecvMsg(m interface{}) error {
	err := w.ClientStream.RecvMsg(m)
	if !w.printed {
		w.printServerAddr()
		w.printed = true
	}
	return err
}

func (w *wrappedStream) SendMsg(m interface{}) error {
	err := w.ClientStream.SendMsg(m)
	if !w.printed {
		w.printServerAddr()
		w.printed = true
	}
	return err
}

func (w *wrappedStream) printServerAddr() {
	if p, ok := peer.FromContext(w.ClientStream.Context()); ok && p.Addr != nil {
		fmt.Printf("Stream call: %s -> Server Address: %s\n", w.method, p.Addr.String())
	}
}

func streamInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	stream, err := streamer(ctx, desc, cc, method, opts...)
	if err != nil {
		return nil, err
	}
	return &wrappedStream{
		ClientStream: stream,
		method:       method,
		printed:      false,
	}, nil
}

func main() {

	// 初始化 zap logger
	//zapLogger := zap.ZapInterceptor()
	zap.ZapInterceptor()
	defer zap.CloseLogger()
	ctx := context.Background()
	cleanup := clienttrace.InitTracer(ctx)
	defer cleanup()

	ctx = clienttrace.FuncCall(ctx, "main")

	conn, err := initClient(ctx)
	if err != nil {
		log.Fatalf("failed to initialize client: %v", err)
	}
	defer conn.Close()

	mux := router.RegisterRoutes(&router.Clients{
		UserClient:       uGrpcClient,
		StreamClient:     sGrpcClient,
		UploadClient:     cGrpcClient,
		BothStreamClient: bothGrpcClient,
		GoodClient:       gGrpcClient,
	})

	if err := http.ListenAndServe(":8081", mux); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}

}

func initClient(ctx context.Context) (*grpc.ClientConn, error) {

	tokens, err := auth.CreateAuth()
	if err != nil {
		return nil, fmt.Errorf("CreateAuth err: %w", err)
	}

	b, _ := json.Marshal(tokens)

	zap.CtxInfof(ctx, "tokens: %v", string(b))

	authToken := auth.Token{
		Value:   "bearer " + tokens.AccessToken,
		TraceId: "123456",
	}

	//etcd服务发现
	r := etcd.NewServerDiscovery(EtcdEndpoints)
	resolver.Register(r)

	//fmt.Println(r.Scheme())

	// 设置默认服务配置（使用 round_robin 负载均衡）
	defaultServiceConfig := `{"loadBalancingPolicy":"round_robin"}`

	//连接服务器
	connect, err := grpc.NewClient(r.Scheme()+":///grpc-demo",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(&authToken),
		grpc.WithDefaultServiceConfig(defaultServiceConfig),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()), //写入traceid到
		grpc.WithUnaryInterceptor(unaryInterceptor),
		grpc.WithStreamInterceptor(streamInterceptor),

		//拦截器注册日志，但是会记录很多系统日志
		//grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(
		//	unaryInterceptor,
		//	grpc_zap.UnaryClientInterceptor(zapLogger),
		//)),
		//grpc.WithStreamInterceptor(grpc_middleware.ChainStreamClient(
		//	streamInterceptor,
		//	grpc_zap.StreamClientInterceptor(zapLogger),
		//)),
	)
	if err != nil {
		return nil, fmt.Errorf("grpc.NewClient err: %w", err)
	}

	// 建立gRPC连接
	uGrpcClient = pb.NewUserServiceClient(connect)
	sGrpcClient = pb.NewStreamServiceClient(connect)
	cGrpcClient = pb.NewStreamClientClient(connect)
	bothGrpcClient = pb.NewStreamClient(connect)

	gGrpcClient = pb.NewGoodClient(connect)

	return connect, nil
}
