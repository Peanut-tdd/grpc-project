package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pbuser/client/utils"
	common "github.com/pbuser/genproto/common"
	pb "github.com/pbuser/genproto/user"
	"net/http"
)

func UserHandler(w http.ResponseWriter, r *http.Request, client pb.UserServiceClient) {
	results, err := UserService(client)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.WriteJSON(w, http.StatusOK, "success", results)
}

type UserResult struct {
	CreateUser *common.CreateUserResponse `json:"create_user"`
	GetUser    *common.GetUserResponse    `json:"get_user"`
}

// 一元rpc
func UserService(client pb.UserServiceClient) (*UserResult, error) {
	createResp, err := client.CreateUser(context.Background(), &common.CreateUserRequest{
		Name:    "John Doe",
		Phone:   "13800138000",
		Address: "XXXXXXX",
		Passwd:  "123456",
		Page:    []int32{1, 2, 3, 4, 5, 6, 7, 8, 9},
	})
	if err != nil {
		return nil, fmt.Errorf("fail to create user: %w", err)
	}

	r, _ := json.Marshal(createResp)
	fmt.Printf("create user resp:%v\n", string(r))

	getResp, err := client.GetUserInfo(context.Background(), &common.GetUserInfoRequest{
		Name:  "John Doe",
		Phone: "13800138000",
	})
	if err != nil {
		return nil, fmt.Errorf("fail to get user info: %w", err)
	}

	r1, _ := json.Marshal(getResp)
	fmt.Printf("get user resp:%v\n", string(r1))

	return &UserResult{
		CreateUser: createResp,
		GetUser:    getResp,
	}, nil
}
