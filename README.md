
## grpc demo
```aiignore
包含了一元，服务端流式，客户端流式，双向流式rpc，集成jwt认证
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