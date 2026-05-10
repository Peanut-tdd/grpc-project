package etcd

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/resolver"
	"log"
	"sync"
	"time"
)

const schema = "grpclb"

type ServerDiscovery struct {
	cli        *clientv3.Client
	cc         resolver.ClientConn
	serverList map[string]resolver.Address
	lock       sync.RWMutex
}

func NewServerDiscovery(endpoints []string) resolver.Builder {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		log.Fatal(err)

	}
	return &ServerDiscovery{
		cli: cli,
	}
}

// Build 为给定目标创建一个新的`resolver`，当调用`grpc.Dial()`时执行
func (s *ServerDiscovery) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	s.cc = cc
	s.serverList = make(map[string]resolver.Address)
	prefix := "/" + target.URL.Scheme + "/" + target.Endpoint() + "/"
	fmt.Println(prefix)

	//根据前缀获取现有的key
	resp, err := s.cli.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		panic(err)
		return nil, err
	}
	for _, kv := range resp.Kvs {
		s.SetServiceList(string(kv.Key), string(kv.Value))
		fmt.Printf("find etcd key:%s, value:%s\n", kv.Key, string(kv.Value))
	}
	s.cc.NewAddress(s.getServiceList())

	go s.watcher(prefix)

	return s, nil
}

func (s *ServerDiscovery) Scheme() string {
	return schema
}

// ResolveNow 监视目标更新
func (s *ServerDiscovery) ResolveNow(resolver.ResolveNowOptions) {
	fmt.Println("ResolveNow")
}

func (s *ServerDiscovery) Close() {
	s.cli.Close()
}

// watcher 监听前缀
func (s *ServerDiscovery) watcher(prefix string) {
	rch := s.cli.Watch(context.Background(), prefix, clientv3.WithPrefix())
	fmt.Println("watcher prefix:", prefix)

	for wresp := range rch {
		for _, ev := range wresp.Events {
			if ev.Type == clientv3.EventTypePut {
				s.SetServiceList(string(ev.Kv.Key), string(ev.Kv.Value))
			}
			if ev.Type == clientv3.EventTypeDelete {
				s.DeleteServiceList(string(ev.Kv.Key))
			}
		}
	}
}

func (s *ServerDiscovery) SetServiceList(key, val string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.serverList[key] = resolver.Address{Addr: val}
	s.cc.NewAddress(s.getServiceList())
	fmt.Println("put key:", key, "val:", val)
}

func (s *ServerDiscovery) DeleteServiceList(key string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.serverList, key)
	s.cc.NewAddress(s.getServiceList())
	fmt.Println("delete key:", key)
}

func (s *ServerDiscovery) getServiceList() []resolver.Address {
	addrs := make([]resolver.Address, 0, len(s.serverList))

	for _, v := range s.serverList {
		addrs = append(addrs, v)
	}

	return addrs
}
