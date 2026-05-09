package service

import (
	"context"
	"fmt"
	"github.com/pbuser/genproto/common"
	pb "github.com/pbuser/genproto/user"
)

type GoodService struct {
	pb.UnimplementedGoodServer
}

func NewGoodService() *GoodService {
	return &GoodService{}
}

func (g *GoodService) CreateGood(ctx context.Context, in *common.CreateGoodReq) (*common.CreateGoodResp, error) {

	good := in.Good
	tags := in.Tags

	fmt.Printf("CreateGood req tag:%+v\n", tags)

	var name, sku, price string

	if val, ok := good["name"]; ok {
		name = val
	}
	if val, ok := good["sku"]; ok {
		sku = val
	}
	if val, ok := good["price"]; ok {
		price = val
	}

	return &common.CreateGoodResp{
		Result: map[string]string{
			"name":  name,
			"sku":   sku,
			"price": price,
		},
		Code: "200",
		Tags: tags,
	}, nil

}
