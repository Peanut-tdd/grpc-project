package gateway

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	pb "github.com/pbuser/genproto/user"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

func ProvideHTTP(grpcAddr string, httpAddr string, grpcServer *grpc.Server) *http.Server {
	ctx := context.Background()

	// grpc服务地址

	gwmux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	// HTTP转grpc
	err := pb.RegisterUserServiceHandlerFromEndpoint(ctx, gwmux, grpcAddr, opts)
	if err != nil {
		grpclog.Fatalf("Register handler err:%v\n", err)
	}

	mux := http.NewServeMux()
	//注册gwmux
	mux.Handle("/", gwmux)
	log.Println(httpAddr + " HTTP.Listing with token...")
	return &http.Server{
		Addr:    httpAddr,
		Handler: grpcHandlerFunc(grpcServer, mux),
	}
}

func grpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	return h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			otherHandler.ServeHTTP(w, r)
		}
	}), &http2.Server{})
}
