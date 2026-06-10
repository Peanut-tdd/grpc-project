package service

import (
	"context"
	"github.com/pbuser/genproto/common"
	pb "github.com/pbuser/genproto/user"
	"log"
	"net/http"
	"strconv"
)

func UploadHandler(w http.ResponseWriter, r *http.Request, client pb.StreamClientClient) {
	Upload(client)
}

// 客户端流式
func Upload(client pb.StreamClientClient) {
	stream, err := client.Upload(context.Background())
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
