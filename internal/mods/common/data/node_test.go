package data

import (
	"context"
	"github.com/stretchr/testify/assert"
	clientv3 "go.etcd.io/etcd/client/v3"
	"testing"
)

func TestNodeRepository_GetNodeInfo(t *testing.T) {
}

func TestEtcdWatcherDelete(t *testing.T) {
	etcdClient, err := clientv3.New(clientv3.Config{Endpoints: []string{"127.0.0.1:2379"}, Username: "root", Password: "123456"})
	assert.Equal(t, err, nil)

	leaseResp, err := etcdClient.Grant(context.Background(), 10)
	assert.Equal(t, err, nil)

	_, err = etcdClient.Put(context.Background(), "task_key", "task_value", clientv3.WithLease(leaseResp.ID))
	assert.Equal(t, err, nil)

	t.Log("already put.. ")

	watchRespChan := etcdClient.Watch(context.Background(), "task_key", clientv3.WithPrefix(), clientv3.WithPrevKV())
	for watchResp := range watchRespChan {
		for _, event := range watchResp.Events {
			if event.Type == clientv3.EventTypeDelete {
				t.Log("delete key: ", string(event.Kv.Key))
			}
		}
	}
}
