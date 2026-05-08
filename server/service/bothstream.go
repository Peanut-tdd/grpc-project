package service

import (
	"fmt"
	"io"
	"strconv"

	"github.com/pbuser/genproto/common"
	pb "github.com/pbuser/genproto/user"
)

type BothStreamServer struct {
	pb.UnimplementedStreamServer
}

func NewBothStreamServer() *BothStreamServer {
	return &BothStreamServer{}
}

func (c *BothStreamServer) Conversations(srv pb.Stream_ConversationsServer) error {
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

		fmt.Printf("conversations result: %s\n", req.Question)
	}
}
