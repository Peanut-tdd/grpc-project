package service

import (
	"io"
	"strconv"

	"github.com/pbuser/genproto/common"
	pb "github.com/pbuser/genproto/user"
	"github.com/pbuser/server/middleware"
	servertrace "github.com/pbuser/server/trace"
)

type BothStreamServer struct {
	pb.UnimplementedStreamServer
}

func NewBothStreamServer() *BothStreamServer {
	return &BothStreamServer{}
}

func (c *BothStreamServer) Conversations(srv pb.Stream_ConversationsServer) error {

	ctx := servertrace.FuncCall(srv.Context(), "Conversations")

	user, err := middleware.GetUserInfo(ctx)
	if err != nil {
		return err
	}

	middleware.CtxInfof(ctx, "user info:%v", user)

	n := 1
	for {
		req, err := srv.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		err = srv.Send(&common.StreamResp{
			Answer: "from stream server answer: the " + strconv.Itoa(n) + " question is " + req.Question,
		})
		if err != nil {
			return err
		}
		n++

		middleware.CtxInfof(ctx, "conversations result: %s", req.Question)
	}
}
