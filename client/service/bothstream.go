package service

import (
	"context"
	"fmt"
	clienttrace "github.com/pbuser/client/trace"
	"github.com/pbuser/client/zap"
	"github.com/pbuser/genproto/common"
	pb "github.com/pbuser/genproto/user"
	"google.golang.org/grpc/grpclog"
	"io"
	"net/http"
	"strconv"
)

func BothStreamHandler(w http.ResponseWriter, r *http.Request, client pb.StreamClient) {
	conversations(client)
}

// 双向流式
func conversations(client pb.StreamClient) {

	ctx := context.Background()
	ctx = clienttrace.FuncCall(ctx, "conversations")

	stream, err := client.Conversations(ctx)

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
