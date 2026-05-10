package etcd

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
	"time"
)

const schema = "grpclb"

type ServiceRegister struct {
	cli           *clientv3.Client
	leaseID       clientv3.LeaseID
	keepAliveChan <-chan *clientv3.LeaseKeepAliveResponse
	key           string
	value         string
}

func NewServiceRegister(endpoints []string, serName, addr string, lease int64) (*ServiceRegister, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: time.Second * 5,
	})

	fmt.Println(err)
	if err != nil {
		panic(err)
		return nil, err
	}

	ser := &ServiceRegister{
		cli:   client,
		key:   "/" + schema + "/" + serName + "/" + addr,
		value: addr,
	}
	//创建租约，注册服务绑定租约，keepalive续约
	err = ser.putKeyWithLease(lease)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return ser, nil
}

func (s *ServiceRegister) putKeyWithLease(lease int64) error {

	//创建带过期时间租约
	resp, err := s.cli.Grant(context.Background(), lease)

	if err != nil {
		return err
	}

	//注册服务并绑定租约
	_, err = s.cli.Put(context.Background(), s.key, s.value, clientv3.WithLease(resp.ID))
	if err != nil {
		return err
	}
	//续约
	leaseRespChan, err := s.cli.KeepAlive(context.Background(), resp.ID)
	if err != nil {
		return err
	}
	s.leaseID = resp.ID
	s.keepAliveChan = leaseRespChan

	log.Printf("put key:%s,value:%s with lease:%v success!", s.key, s.value, s.leaseID)

	return nil
}

func (s *ServiceRegister) ListenLeaseRespChan() {
	for leaseKeepResp := range s.keepAliveChan {

		fmt.Println("续租成功", leaseKeepResp)
	}

}
func (s *ServiceRegister) Close() error {
	_, err := s.cli.Revoke(context.Background(), s.leaseID)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	fmt.Println("撤销续租")
	return s.cli.Close()
}
