package service

import (
	"context"
	"fmt"
	"github.com/pbuser/genproto/common"
	pb "github.com/pbuser/genproto/user"
	"google.golang.org/grpc/grpclog"
	"io"
	"net/http"
)

func StreamHandler(w http.ResponseWriter, r *http.Request, client pb.StreamServiceClient) {
	ListValue(client)
}

// 服务端流式
func ListValue(client pb.StreamServiceClient) {

	req := &common.SimpleRequest{
		Data: "stream server start...",
	}

	stream, err := client.ListValue(context.Background(), req)
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
