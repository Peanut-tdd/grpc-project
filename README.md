
## grpc demo
```aiignore
包含了一元，服务端流式，客户端流式，双向流式rpc，集成jwt认证,zap日志
```

#### 目录结构
```aiignore
.
├── README.md
├── client
│   ├── auth                    //jwt认证
│   │   └── auth.go
│   ├── client.go
│   ├── etcd                        //etcd服务发现与watch
│   │   └── discovery.go
│   ├── trace                       //链路追踪
│   │   └── trace.go
│   └── zap                        //zap日志
│       └── zap.go                  
├── genproto                        //pb.go文件目录
│   ├── common
│   │   └── common.pb.go
│   └── user
│       ├── user.pb.go
│       ├── user.pb.gw.go
│       └── user_grpc.pb.go
├── go.mod
├── go.sum
├── img
│   ├── img.png
│   ├── img_1.png
│   ├── img_2.png
│   ├── img_4.png
│   └── img_5.png
├── proto                           //pb定义文件目录
│   ├── common
│   │   └── common.proto
│   ├── google
│   │   └── api
│   │       ├── annotations.proto
│   │       └── http.proto
│   └── user
│       └── user.proto          
└── server
    ├── jwt
    │   └── jwt.go                 //jwt
    ├── middleware
    │   ├── auth.go             //jwt认证
    │   ├── cancel.go           //超时控制
    │   └── zap.go              //zap日志
    ├── server.go                          
    ├── service                      
    │   ├── bothstream.go       //双向流式rpc
    │   ├── etcd                //etcd服务注册
    │   │   └── register.go
    │   ├── gateway                 //rpc服务转http，实现支持curl请求
    │   │   └── gateway.go
    │   ├── good.go
    │   ├── stream.go       //服务端流式rpc
    │   ├── upload.go       //客户端流式rpc
    │   └── user.go         //一元rpc
    └── trace          //创建 tracer Provider
        └── trace.go


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


2.添加路由配置
![img.png](img/img.png)

3.生成下载annotations.pb.go和http.pb.go文件
```aiignore
protoc -I./proto --go_out=./genproto    --go_opt=paths=source_relative ./proto/api/*.proto
```

4.生成pb.gw.go文件
```aiignore
protoc -I./proto \
  --go_out=./genproto --go_opt=paths=source_relative \
  --go-grpc_out=./genproto --go-grpc_opt=paths=source_relative \
  --grpc-gateway_out=./genproto --grpc-gateway_opt=paths=source_relative \
  ./proto/user/*.proto
```

5.在service目录下添加gateway目录，创建gateway.go


6.在server.go文件中添加Http注册
```aiignore
httpServer := gateway.ProvideHTTP("127.0.0.1"+Addr, "127.0.0.1"+HttpAddr, grpcServer)
fmt.Printf("HTTP Gateway listening on %s\n", fmt.Sprintf("%s%s", "127.0.0.1", HttpAddr))
if err = httpServer.ListenAndServe(); err != nil {
    log.Fatal("ListenAndServe: ", err)
}
```



7.postman 携带bearer token请求
![img_1.png](img/img_1.png)








#### 使用otel和jaeger完成客户端到服务端的链路追踪和上报，同时将traceid写入客户端&服务端zap日志中

在客户端&服务端先初始化tracer

客户端
```aiignore

cleanup := clienttrace.InitTracer(ctx)
defer cleanup()
```

服务端
```aiignore
cleanup := servertrace.InitTracer(ctx)
defer cleanup()
```



在客户端拦截器添加如下代码，将traceid写入metadata中
```aiignore

grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
```
在客户端拦截器添加如下代码，自动解析metadata中的traceid，并创建一个子 span 继承父 span 的 trace_id
```aiignore
grpc.StatsHandler(otelgrpc.NewServerHandler())
```


![img_2.png](img/img_2.png)

客户端带traceid日志
![img_4.png](img/img_4.png)

服务端带traceid日志
![img_5.png](img/img_5.png)





