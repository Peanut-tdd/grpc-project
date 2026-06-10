package main

import (
	"context"
	"github.com/pbuser/client/router"
	clienttrace "github.com/pbuser/client/trace"
	"github.com/pbuser/client/zap"
	"github.com/pbuser/svc"
	"log"
	"net/http"
)

var EtcdEndpoints = []string{"localhost:2379"}

func main() {

	// 初始化 zap logger
	//zapLogger := zap.ZapInterceptor()
	zap.ZapInterceptor()
	defer zap.CloseLogger()
	ctx := context.Background()
	cleanup := clienttrace.InitTracer(ctx)
	defer cleanup()

	ctx = clienttrace.FuncCall(ctx, "main")

	//服务发现，指定需要认证的服务（不传则所有服务都需要认证）
	svcCtx, err := svc.NewServiceContext(ctx, EtcdEndpoints)
	if err != nil {
		log.Fatalf("failed to initialize service context: %v", err)
	}
	defer svcCtx.Close()

	//路由
	mux := router.RegisterRoutes(&router.Clients{
		UserClient:       svcCtx.UserClient,
		StreamClient:     svcCtx.StreamClient,
		UploadClient:     svcCtx.UploadClient,
		BothStreamClient: svcCtx.BothStreamClient,
		GoodClient:       svcCtx.GoodClient,
	})

	if err := http.ListenAndServe(":8081", mux); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}

}
