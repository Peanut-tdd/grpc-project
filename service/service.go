package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pbuser/genproto/common"
	pb "github.com/pbuser/genproto/user"
	"google.golang.org/grpc"
	"io"
	"log"
	"net"
	"strconv"
	"time"
)

type userService struct {
	pb.UnimplementedUserServiceServer
}

func NewUserService() pb.UserServiceServer {
	return &userService{}
}

func (u userService) CreateUser(ctx context.Context, rq *common.CreateUserRequest) (*common.CreateUserResponse, error) {

	str, _ := json.Marshal(rq)
	fmt.Println(string(str))

	return &common.CreateUserResponse{
		Id:    1,
		Name:  rq.Name,
		Phone: rq.Phone,
	}, nil
}

func (u userService) GetUserInfo(ctx context.Context, rq *common.GetUserInfoRequest) (*common.GetUserResponse, error) {

	str, _ := json.Marshal(rq)
	fmt.Println(string(str))
	return &common.GetUserResponse{
		Id:    1,
		Name:  "tdd520",
		Phone: "16666666",
	}, nil

}

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

const (
	Addr    = ":8080"
	NetWork = "tcp"
)

func main() {

	lister, err := net.Listen(NetWork, Addr)
	if err != nil {
		log.Fatalf("net.Listen err: %v", err)
		return
	}

	defer lister.Close()

	fmt.Println("server lister is ", lister.Addr())

	grpcServer := grpc.NewServer()
	pb.RegisterUserServiceServer(grpcServer, NewUserService())
	pb.RegisterStreamServiceServer(grpcServer, NewStreamService())
	pb.RegisterStreamClientServer(grpcServer, NewUploadService())
	pb.RegisterStreamServer(grpcServer, NewBothStreamServer())

	err = grpcServer.Serve(lister)
	if err != nil {
		log.Fatalf("grpcServer.Serve err: %v", err)
	}
}
