package registry

import (
	"context"
	v1 "github.com/peckfly/gopeck/api/pecker/v1"
	"github.com/peckfly/gopeck/internal/mods/pecker/service"
	"github.com/stretchr/testify/assert"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
	"testing"
)

func TestServer(t *testing.T) {
	etcdClient, err := clientv3.New(clientv3.Config{Endpoints: []string{"127.0.0.1:2379"}, Username: "root", Password: "123456"})
	assert.Equal(t, err, nil)
	etcdRegistry := NewEtcdRegistry(etcdClient)

	serviceName := "test"

	srv := NewGrpcServer(serviceName, "127.0.0.1:8082", WithServerRegistrar(etcdRegistry))
	v1.RegisterPeckServiceServer(srv, &service.PeckService{})
	err = srv.Start()
	assert.Equal(t, err, nil)
}

func TestKeepAlive(t *testing.T) {
	cli, err := clientv3.New(clientv3.Config{Endpoints: []string{"127.0.0.1:2379"}, Username: "root", Password: "123456"})
	resp, err := cli.Grant(context.Background(), 10)
	if err != nil {
		log.Fatal(err)
	}
	leaseID := resp.ID

	ch, kErr := cli.KeepAlive(context.Background(), leaseID)
	if kErr != nil {
		log.Fatal(kErr)
	}

	for {
		select {
		case ka := <-ch:
			log.Printf("ttl: %d", ka.TTL)
		}
	}

}
