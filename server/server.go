package main

import (
	"fmt"
	"github.com/pbuser/server/service"
	"log"
	"net"

	pb "github.com/pbuser/genproto/user"
	"google.golang.org/grpc"
)

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
	pb.RegisterUserServiceServer(grpcServer, service.NewUserService())
	pb.RegisterStreamServiceServer(grpcServer, service.NewStreamService())
	pb.RegisterStreamClientServer(grpcServer, service.NewUploadService())
	pb.RegisterStreamServer(grpcServer, service.NewBothStreamServer())

	err = grpcServer.Serve(lister)
	if err != nil {
		log.Fatalf("grpcServer.Serve err: %v", err)
	}
}
