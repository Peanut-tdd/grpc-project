package svc

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pbuser/client/auth"
	"github.com/pbuser/client/etcd"
	pb "github.com/pbuser/genproto/user"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/resolver"
)

// ServiceContext 聚合 gRPC 连接及所有服务客户端

var AuthServices = []string{"UserService", "StreamService", "StreamClient", "Stream"}

type ServiceContext struct {
	Conn *grpc.ClientConn

	UserClient       pb.UserServiceClient
	StreamClient     pb.StreamServiceClient
	UploadClient     pb.StreamClientClient
	BothStreamClient pb.StreamClient
	GoodClient       pb.GoodClient
}

// selectiveAuth 按服务名选择性注入认证信息
type selectiveAuth struct {
	token   string
	traceId string
	// services 需要认证的服务名集合，key 为服务名（如 "UserService"）
	// 为空时所有服务都需要认证
	services map[string]bool
}

func (s *selectiveAuth) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	// 未配置特定服务时，所有服务都需要认证
	if len(s.services) == 0 {
		return map[string]string{
			"authorization": s.token,
			"traceId":       s.traceId,
		}, nil
	}

	// 根据请求 URI 判断是否需要认证
	// uri[0] 格式: "/user.UserService/CreateUser"
	if len(uri) > 0 && s.matchService(uri[0]) {
		return map[string]string{
			"authorization": s.token,
			"traceId":       s.traceId,
		}, nil
	}

	// 不需要认证的服务，不注入 authorization
	return nil, nil
}

// matchService 检查 RPC URI 对应的服务是否需要认证
func (s *selectiveAuth) matchService(uri string) bool {
	for svc := range s.services {
		if strings.Contains(uri, svc) {
			return true
		}
	}
	return false
}

func (s *selectiveAuth) RequireTransportSecurity() bool {
	return false
}

var _ credentials.PerRPCCredentials = (*selectiveAuth)(nil)

// NewServiceContext 初始化服务发现并创建 gRPC 连接与客户端
// authServices: 需要认证的服务名列表（如 "UserService", "Stream"），为空则所有服务都需要认证
func NewServiceContext(ctx context.Context, etcdEndpoints []string, authServices ...string) (*ServiceContext, error) {
	tokens, err := auth.CreateAuth()
	if err != nil {
		return nil, fmt.Errorf("CreateAuth err: %w", err)
	}

	b, _ := json.Marshal(tokens)
	fmt.Printf("tokens: %v\n", string(b))

	// 构建需要认证的服务集合
	svcSet := make(map[string]bool, len(AuthServices))
	for _, s := range AuthServices {
		svcSet[s] = true
	}

	authCred := &selectiveAuth{
		token:    "bearer " + tokens.AccessToken,
		traceId:  "123456",
		services: svcSet,
	}

	// etcd 服务发现
	r := etcd.NewServerDiscovery(etcdEndpoints)
	resolver.Register(r)

	// round_robin 负载均衡
	defaultServiceConfig := `{"loadBalancingPolicy":"round_robin"}`

	conn, err := grpc.NewClient(r.Scheme()+":///grpc-demo",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(authCred),
		grpc.WithDefaultServiceConfig(defaultServiceConfig),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithUnaryInterceptor(unaryInterceptor),
		grpc.WithStreamInterceptor(streamInterceptor),
	)
	if err != nil {
		return nil, fmt.Errorf("grpc.NewClient err: %w", err)
	}

	return &ServiceContext{
		Conn:             conn,
		UserClient:       pb.NewUserServiceClient(conn),
		StreamClient:     pb.NewStreamServiceClient(conn),
		UploadClient:     pb.NewStreamClientClient(conn),
		BothStreamClient: pb.NewStreamClient(conn),
		GoodClient:       pb.NewGoodClient(conn),
	}, nil
}

// Close 关闭 gRPC 连接
func (s *ServiceContext) Close() error {
	return s.Conn.Close()
}

func unaryInterceptor(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
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

func (w *wrappedStream) RecvMsg(m any) error {
	err := w.ClientStream.RecvMsg(m)
	if !w.printed {
		w.printServerAddr()
		w.printed = true
	}
	return err
}

func (w *wrappedStream) SendMsg(m any) error {
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
