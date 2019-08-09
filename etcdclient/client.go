package etcdclient

import (
	"consual/common"
	"context"
	"fmt"
	"time"

	client "go.etcd.io/etcd/clientv3"
)

type EtcdClient struct {
	etcdConn *client.Client
	Kv       client.KV
}

func NewEtcdClient() (*EtcdClient, error) {
	config := client.Config{
		Endpoints:   []string{"192.168.205.10:2379", "192.168.205.11:2379"},
		DialTimeout: time.Second * 10,
	}

	etcdConn, err := client.New(config)

	if err != nil {
		return nil, err
	}

	kv := client.NewKV(etcdConn)
	return &EtcdClient{etcdConn: etcdConn, Kv: kv}, nil
}

func (this *EtcdClient) RegisterService(id, name, address string) error {
	//设置租约
	lease := client.NewLease(this.etcdConn)
	grantLease, err := lease.Grant(context.Background(), 20)

	if err != nil {
		return err
	}
	_, err = this.Kv.Put(context.Background(), common.ServicePrefix+id+"/"+name, address, client.WithLease(grantLease.ID))
	if err != nil {
		return err
	}

	keepAlive, err := lease.KeepAlive(context.TODO(), grantLease.ID)

	if err != nil {
		return err
	}

	go keepAliveTry(keepAlive)

	return nil

}

func keepAliveTry(keepAliveResp <-chan *client.LeaseKeepAliveResponse) {
	for {
		select {
		case ret := <-keepAliveResp:
			if ret != nil {
				fmt.Println("keep alive ....")
			}
		}
	}
}
func (this *EtcdClient) UnRegisterService(id, name string) error {
	_, err := this.Kv.Delete(context.Background(), common.ServicePrefix+id, client.WithPrefix())
	return err
}
