package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strconv"

	"github.com/pbuser/client/auth"
	common "github.com/pbuser/genproto/common"
	pb "github.com/pbuser/genproto/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

const Addr = ":8080"

var (
	uGrpcClient    pb.UserServiceClient
	sGrpcClient    pb.StreamServiceClient
	cGrpcClient    pb.StreamClientClient
	bothGrpcClient pb.StreamClient
)

func main() {
	conn, err := initClient()
	if err != nil {
		log.Fatalf("failed to initialize client: %v", err)
	}
	defer conn.Close()

	//UserService()
	//ListValue()

	//Upload()
	conversations()
}

func initClient() (*grpc.ClientConn, error) {
	tokens, err := auth.CreateAuth()
	if err != nil {
		return nil, fmt.Errorf("CreateAuth err: %w", err)
	}

	b, _ := json.Marshal(tokens)
	fmt.Println(string(b))

	authToken := auth.Token{
		Value: "bearer " + tokens.AccessToken,
	}

	//连接服务器
	connect, err := grpc.Dial(Addr, grpc.WithInsecure(), grpc.WithPerRPCCredentials(&authToken))
	if err != nil {
		return nil, fmt.Errorf("grpc.Dial err: %w", err)
	}

	// 建立gRPC连接
	uGrpcClient = pb.NewUserServiceClient(connect)
	sGrpcClient = pb.NewStreamServiceClient(connect)
	cGrpcClient = pb.NewStreamClientClient(connect)
	bothGrpcClient = pb.NewStreamClient(connect)

	return connect, nil
}

// 一元rpc
func UserService() {
	resp, err := uGrpcClient.CreateUser(context.Background(), &common.CreateUserRequest{
		Name:    "John Doe",
		Phone:   "13800138000",
		Address: "XXXXXXX",
		Passwd:  "123456",
	})
	if err != nil {
		grpclog.Fatalf("fail to create user: %v", err)
	}

	r, _ := json.Marshal(resp)
	fmt.Printf("create user resp:%v\n", string(r))

	resp1, err := uGrpcClient.GetUserInfo(context.Background(), &common.GetUserInfoRequest{
		Name:  "John Doe",
		Phone: "13800138000",
	})
	if err != nil {
		grpclog.Fatalf("fail to get user info: %v", err)
	}

	r1, _ := json.Marshal(resp1)
	fmt.Printf("get user resp:%v\n", string(r1))
}

// 服务端流式
func ListValue() {

	req := &common.SimpleRequest{
		Data: "stream server start...",
	}

	stream, err := sGrpcClient.ListValue(context.Background(), req)
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

// 客户端流式
func Upload() {
	stream, err := cGrpcClient.Upload(context.Background())
	if err != nil {
		log.Fatalf("fail to upload: %v", err)
	}

	for i := 0; i < 10; i++ {
		err = stream.Send(&common.StreamRequest{StreamData: "stream client rpc" + strconv.Itoa(i)})
		if err != nil {
			log.Fatalf("fail to upload: %v", err)
		}
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("fail to upload: %v", err)
	}

	log.Printf("upload resp:%v\n", res)

}

// 双向流式
func conversations() {

	stream, err := bothGrpcClient.Conversations(context.Background())

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

	}

	err = stream.CloseSend()
	if err != nil {
		grpclog.Fatalf("fail to conversations: %v", err)
	}

}
