
## grpc demo
```aiignore
包含了一元，服务端流式，客户端流式，双向流式rpc，集成jwt认证,otel+jaeger链路追踪，zap日志,etcd服务注册发现，负载均衡,pprof性能分析
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


#### etcd注册和发现，负载均衡
先docker-compose部署etcd服务





服务端创建租约+服务绑定租约+keepalive续租
部分代码片段：
![img_12.png](img/img_12.png)

客户端拉取和watcher服务变化并缓存到本地
部分代码片段：
![img_13.png](img/img_13.png)


etcd未启动服务端报错：
![img_14.png](img/img_14.png)



成功注册etcd服务:
![img_15.png](img/img_15.png)




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
在服务端拦截器添加如下代码，自动解析metadata中的traceid，并创建一个子 span 继承父 span 的 trace_id
```aiignore
grpc.StatsHandler(otelgrpc.NewServerHandler())
```


![img_2.png](img/img_2.png)

客户端带traceid日志
![img_4.png](img/img_4.png)

服务端带traceid日志
![img_5.png](img/img_5.png)






#### grpcurl调试
server端添加代码
```aiignore
	reflection.Register(grpcServer)
```

携带auth token 的grpcurl,查看服务列表
```aiignore
 grpcurl -plaintext -H "authorization: bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxMjMsInVzZXJfbmFtZSI6ImFkbWluIiwiZXhwIjoxNzc4NDkyMzA3LCJpYXQiOjE3Nzg0OTE4ODd9.f6H46qgVY6fd8ZRvtUHpEjELXVDlnStlcivhuvUo5sY" 127.0.0.1:8080 list 
```

![img_6.png](img/img_6.png)

查看服务提供的方法
```aiignore
grpcurl -plaintext -H "authorization: bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxMjMsInVzZXJfbmFtZSI6ImFkbWluIiwiZXhwIjoxNzc4NDk1NzYxLCJpYXQiOjE3Nzg0OTM5NjF9.TJdstNmf-FDsCYG8AtWVSIE8I0luUo98IR9p18HaINw" 127.0.0.1:8080 list  user.UserService
```
![img_7.png](img/img_7.png)


查看服务方法详情
```aiignore
grpcurl -plaintext -H "authorization: bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxMjMsInVzZXJfbmFtZSI6ImFkbWluIiwiZXhwIjoxNzc4NDk1NzYxLCJpYXQiOjE3Nzg0OTM5NjF9.TJdstNmf-FDsCYG8AtWVSIE8I0luUo98IR9p18HaINw" 127.0.0.1:8080 describe  user.UserService
```
![img_8.png](img/img_8.png)



#### grpcui调试
携带token的grpcui命令
```aiignore
grpcui -plaintext  -H "authorization: bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxMjMsInVzZXJfbmFtZSI6ImFkbWluIiwiZXhwIjoxNzc4NDk1NzYxLCJpYXQiOjE3Nzg0OTM5NjF9.TJdstNmf-FDsCYG8AtWVSIE8I0luUo98IR9p18HaINw" 127.0.0.1:8080
```
执行后返回可调试的url链接，可直接跳转到对应页面调试
如下图：
![img_9.png](img/img_9.png)






#### postman带auth token调试

![img_10.png](img/img_10.png)






#### ghz压测

安装ghz
```aiignore
brew install ghz
```


进入项目目录，执行压测命令
```aiignore
ghz \
  --proto=proto/user/user.proto \
  -i proto \
  --call=user.UserService/CreateUser \
  --insecure \
  -d '{"name":"tdd","phone":"123456789","address":"asdf","email":"2578741722@qq.com","passwd":"123"}' \
  -m '{"authorization":"Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxMjMsInVzZXJfbmFtZSI6ImFkbWluIiwiZXhwIjoxNzc4NjcwMDA1LCJpYXQiOjE3Nzg2NjgyMDV9.PL5JqM3zwdbGRnQU_Cgwbqw3Lee8wP2PFhj_BXT6i7U"}'\
  -c 1000 \
  -n 100000 \
  -t 30s \
  127.0.0.1:8080

```



压测结果
```aiignore
Summary:
  Count:        100000
  Total:        4.23 s
  Slowest:      248.33 ms
  Fastest:      0.20 ms
  Average:      33.56 ms                #平均响应时长RT
  Requests/sec: 23665.71               #QPS

Response time histogram:            #响应分布
  0.198   [1]     |
  25.011  [27773] |∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
  49.825  [62088] |∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
  74.638  [7593]  |∎∎∎∎∎
  99.452  [1612]  |∎
  124.265 [848]   |∎
  149.079 [72]    |
  173.893 [10]    |
  198.706 [1]     |
  223.520 [1]     |
  248.333 [1]     |

Latency distribution:
  10 % in 15.62 ms 
  25 % in 23.86 ms 
  50 % in 32.61 ms 
  75 % in 40.97 ms 
  90 % in 49.98 ms 
  95 % in 58.99 ms 
  99 % in 97.68 ms 

Status code distribution:
  [OK]   100000 responses 
```





#### pprof的性能分析

添加pprof配置
```aiignore
func startPprof() {
	mux := http.NewServeMux()
	mux.Handle("/debug/", http.DefaultServeMux)

	server := &http.Server{
		Addr:    "0.0.0.0:6060",
		Handler: mux,
	}

	log.Println("pprof server listening on :6060")
	if err := server.ListenAndServe(); err != nil {
		log.Printf("pprof server error: %v", err)
	}
}


```
main函数中调用
```aiignore
	go startPprof()
```

然后，我们在浏览器中访问http://localhost:6060/debug/pprof/，可以看到如下界面：
![img_11.png](img/img_11.png)

其中：
- allocs：所有的历史内存分配情况。
- block：阻塞情况。
- cmdline：命令行调用信息。
- goroutine：协程情况。
- heap：当前活跃对象的堆内存分配采集情况。
- mutex：互斥锁信息。
- profile：cpu使用情况。
- threadcreate：新的操作系统线程创建情况。
- trace：当前程序的执行trace。


安装图可视化工具集，配合pprof web使用
```aiignore
brew install graphviz
```


采集过去20s的cpu使用数据
```aiignore
go tool pprof -seconds 20 -http=localhost:6061 http://localhost:6060/debug/pprof/profile
```


采集过去20s的堆内存使用情况
```aiignore
 go tool pprof -seconds 20 -http=localhost:6061 http://localhost:6060/debug/pprof/heap
```










