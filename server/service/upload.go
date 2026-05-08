package service

import (
	"fmt"
	"github.com/pbuser/genproto/common"
	pb "github.com/pbuser/genproto/user"
	"io"
)

type UploadService struct {
	pb.UnimplementedStreamClientServer
}

func NewUploadService() *UploadService {
	return &UploadService{}
}

func (up *UploadService) Upload(srv pb.StreamClient_UploadServer) error {
	for {
		res, err := srv.Recv()
		if err == io.EOF {
			return srv.SendAndClose(&common.SimpleResponse{Value: "ok"})
		}
		if err != nil {
			return err
		}

		fmt.Printf("upload result: %v\n", res.StreamData)
	}
}
