package service

import (
	"github.com/pbuser/genproto/common"
	pb "github.com/pbuser/genproto/user"
	"strconv"
	"time"
)

type StreamService struct {
	pb.UnimplementedStreamServiceServer
}

func NewStreamService() *StreamService {
	return &StreamService{}
}

func (s *StreamService) ListValue(in *common.SimpleRequest, srv pb.StreamService_ListValueServer) error {
	for i := 0; i < 10; i++ {
		time.Sleep(time.Second)
		err := srv.Send(&common.StreamResponse{
			StreamValue: in.Data + strconv.Itoa(i),
		})
		if err != nil {
			return err
		}
	}
	return nil
}
