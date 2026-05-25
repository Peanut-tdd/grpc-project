package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pbuser/genproto/common"
	pb "github.com/pbuser/genproto/user"
	"github.com/pbuser/server/middleware"
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

	res := &common.CreateUserResponse{
		Id:    1,
		Name:  rq.Name,
		Phone: rq.Phone,
		Page:  rq.Page,
	}

	middleware.CtxInfof(ctx, "user info:%v", res)
	return res, nil
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
