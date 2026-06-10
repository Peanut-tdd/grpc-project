package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pbuser/genproto/common"
	pb "github.com/pbuser/genproto/user"
	"google.golang.org/grpc/grpclog"
	"net/http"
)

func GoodHandler(w http.ResponseWriter, r *http.Request, client pb.GoodClient) {
	CreateGood(client)
}

func CreateGood(client pb.GoodClient) {
	req := &common.CreateGoodReq{
		Good: map[string]string{
			"name":  "Apple 18 ProMax",
			"sku":   "00001",
			"price": "18000",
		},
		Tags: []common.Tags{common.Tags_Tag_Elec, common.Tags_Tag_Hot},
	}

	resp, err := client.CreateGood(context.Background(), req)
	if err != nil {
		grpclog.Fatalf("fail to create good: %v", err)
	}

	js, _ := json.Marshal(resp)
	fmt.Printf("create good resp:%v\n", string(js))
}
