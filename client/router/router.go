package router

import (
	"github.com/pbuser/client/service"
	pb "github.com/pbuser/genproto/user"
	"net/http"
)

// Clients 聚合所有 gRPC 客户端
type Clients struct {
	UserClient       pb.UserServiceClient
	StreamClient     pb.StreamServiceClient
	UploadClient     pb.StreamClientClient
	BothStreamClient pb.StreamClient
	GoodClient       pb.GoodClient
}

// RegisterRoutes 注册所有 HTTP 路由，返回 *http.ServeMux
func RegisterRoutes(c *Clients) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		service.UserHandler(w, r, c.UserClient)
	})

	mux.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
		service.StreamHandler(w, r, c.StreamClient)
	})

	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		service.UploadHandler(w, r, c.UploadClient)
	})

	mux.HandleFunc("/both", func(w http.ResponseWriter, r *http.Request) {
		service.BothStreamHandler(w, r, c.BothStreamClient)
	})

	mux.HandleFunc("/good", func(w http.ResponseWriter, r *http.Request) {
		service.GoodHandler(w, r, c.GoodClient)
	})

	return mux
}
