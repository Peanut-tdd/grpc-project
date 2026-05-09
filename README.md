
## grpc demo
```aiignore
包含了一元，服务端流式，客户端流式，双向流式rpc，集成jwt认证,zap日志
```

#### 目录结构
```aiignore
├.
├── README.md
├── client                                //client端
│   ├── auth
│   │   └── auth.go
│   └── client.go
├── genproto                            //pb文件生成目录
│   ├── common
│   │   └── common.pb.go
│   └── user
│       ├── user.pb.go
│       └── user_grpc.pb.go
├── go.mod
├── go.sum
├── proto                           //proto定义文件目录
│   ├── common
│   │   └── common.proto   
│   └── user
│       └── user.proto
└── server
    ├── jwt                     //jwt 认证
    │   └── jwt.go
    ├── middleware
    │   └── auth.go         //auth验证
    ├── server.go
    └── service
        ├── bothstream.go       //双向流式
        ├── stream.go           //服务端流式
        ├── upload.go           //客户端流式
        └── user.go             //普通rpc

```




#### 生成pb文件到指定目录
```aiignore
protoc -I./proto --go_out=./genproto --go_opt  paths=source_relative --go-grpc_out=./genproto  --go-grpc_opt=paths=source_relative ./proto/user/*.proto
```






#### rpc服务转http

##### 下载grpc-gateway组件
```aiignore
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
```

1.下载annotations.proto 和 http.proto 到本地[下载源](https://github.com/googleapis/googleapis/tree/master/google/api)

2.生成下载annotations.pb.go和http.pb.go文件
```aiignore
protoc -I./proto --go_out=./genproto    --go_opt=paths=source_relative ./proto/api/*.proto
```

3.生成pb.gw.go文件
```aiignore
protoc -I./proto \
  --go_out=./genproto --go_opt=paths=source_relative \
  --go-grpc_out=./genproto --go-grpc_opt=paths=source_relative \
  --grpc-gateway_out=./genproto --grpc-gateway_opt=paths=source_relative \
  ./proto/user/*.proto
```

4.在service目录下添加gateway目录，创建gateway.go


5.在server.go文件中添加Http注册
```aiignore
httpServer := gateway.ProvideHTTP("127.0.0.1"+Addr, "127.0.0.1"+HttpAddr, grpcServer)
fmt.Printf("HTTP Gateway listening on %s\n", fmt.Sprintf("%s%s", "127.0.0.1", HttpAddr))
if err = httpServer.ListenAndServe(); err != nil {
    log.Fatal("ListenAndServe: ", err)
}
```





