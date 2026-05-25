package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strconv"

	"github.com/pbuser/client/auth"
	"github.com/pbuser/client/etcd"
	clienttrace "github.com/pbuser/client/trace"
	"github.com/pbuser/client/zap"
	common "github.com/pbuser/genproto/common"
	pb "github.com/pbuser/genproto/user"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/resolver"
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

	UserService()
	//ListValue()

	//Upload()
	//conversations(ctx)
	//CreateGood()
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

// 一元rpc
func UserService() {
	resp, err := uGrpcClient.CreateUser(context.Background(), &common.CreateUserRequest{
		Name:    "John Doe",
		Phone:   "13800138000",
		Address: "XXXXXXX",
		Passwd:  "123456",
		Page:    []int32{1, 2, 3, 4, 5, 6, 7, 8, 9},
	})
	if err != nil {
		grpclog.Fatalf("fail to create user: %v", err)
	}

	r, _ := json.Marshal(resp)
	fmt.Printf("create user resp:%v\n", string(r))

	resp1, err := uGrpcClient.GetUserInfo(context.Background(), &common.GetUserInfoRequest{
		Name:  "John Doe",
		Phone: "13800138000",
	})
	if err != nil {
		grpclog.Fatalf("fail to get user info: %v", err)
	}

	r1, _ := json.Marshal(resp1)
	fmt.Printf("get user resp:%v\n", string(r1))
}

// 服务端流式
func ListValue() {

	req := &common.SimpleRequest{
		Data: "stream server start...",
	}

	stream, err := sGrpcClient.ListValue(context.Background(), req)
	if err != nil {
		grpclog.Fatalf("fail to list value: %v", err)
	}

	for {
		//Recv() 方法接收服务端消息，默认每次Recv()最大消息长度为`1024*1024*4`bytes(4M)
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			grpclog.Fatalf("fail to list value: %v", err)
		}
		fmt.Printf("get value resp:%v\n", resp.StreamValue)

	}

	stream.CloseSend()
}

// 客户端流式
func Upload() {
	stream, err := cGrpcClient.Upload(context.Background())
	if err != nil {
		log.Fatalf("fail to upload: %v", err)
	}

	for i := 0; i < 10; i++ {
		err = stream.Send(&common.StreamRequest{StreamData: "stream client rpc" + strconv.Itoa(i)})
		if err != nil {
			log.Fatalf("fail to upload: %v", err)
		}
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("fail to upload: %v", err)
	}

	log.Printf("upload resp:%v\n", res)

}

// 双向流式
func conversations(ctx context.Context) {

	ctx = clienttrace.FuncCall(ctx, "conversations")

	stream, err := bothGrpcClient.Conversations(ctx)

	if err != nil {
		grpclog.Fatalf("fail to conversations: %v", err)
	}

	for i := 0; i < 10; i++ {
		err = stream.Send(&common.StreamReq{
			Question: "stream client rpc" + strconv.Itoa(i),
		})
		if err != nil {
			grpclog.Fatalf("fail to conversations: %v", err)
		}

		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			grpclog.Fatalf("fail to conversations: %v", err)
		}

		fmt.Printf("conversations resp:%v\n", res.Answer)

		zap.CtxInfof(ctx, "conversations resp:%v", res.Answer)

	}

	err = stream.CloseSend()
	if err != nil {
		grpclog.Fatalf("fail to conversations: %v", err)
	}

}

func CreateGood() {
	req := &common.CreateGoodReq{
		Good: map[string]string{
			"name":  "Apple 18 ProMax",
			"sku":   "00001",
			"price": "18000",
		},
		Tags: []common.Tags{common.Tags_Tag_Elec, common.Tags_Tag_Hot},
	}

	resp, err := gGrpcClient.CreateGood(context.Background(), req)
	if err != nil {
		grpclog.Fatalf("fail to create good: %v", err)
	}

	js, _ := json.Marshal(resp)
	fmt.Printf("create good resp:%v\n", string(js))
}
