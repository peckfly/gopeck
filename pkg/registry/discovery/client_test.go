package discovery

import (
	"context"
	"fmt"
	"github.com/peckfly/gopeck/pkg/registry"
	"github.com/stretchr/testify/assert"
	clientv3 "go.etcd.io/etcd/client/v3"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	etcdClient, err := clientv3.New(clientv3.Config{Endpoints: []string{"127.0.0.1:2379"}, Username: "root", Password: "123456"})
	assert.Equal(t, err, nil)
	etcdRegistry := registry.NewEtcdRegistry(etcdClient)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	// service discovery
	serviceName := "gopeck-pecker"
	grpcClientConn, err := Dial(ctx, fmt.Sprintf("%s:///%s", Scheme, serviceName), WithDiscovery(etcdRegistry))
	defer func() {
		_ = grpcClientConn.Close()
	}()
	cancel()

	//userServiceClient := v1.NewPeckServiceClient(grpcClientConn)
	//
	//ctx, cancel = context.WithTimeout(context.Background(), time.Second*20)
	//Peck, err := userServiceClient.Peck(ctx, &v1.PeckRequest{})
	//cancel()

	//t.Log(Peck)
	//assert.Equal(t, err, nil)
}
