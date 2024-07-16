package data

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/peckfly/gopeck/internal/mods/admin/biz"
	"github.com/peckfly/gopeck/pkg/log/logc"
	"github.com/redis/go-redis/v9"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

const (
	taskKey           = "/task/scheduled_key"
	taskLockKey       = "gopeck:scheduled:lock_key"
	delaySecondIfFail = 180 //  if failed, delay 180s to execute again
)

type scheduledTaskRepository struct {
	etcdClient  *clientv3.Client
	redisClient *redis.Client
}

func NewScheduledTaskRepository(etcdClient *clientv3.Client, redisClient *redis.Client) biz.ScheduledTaskRepository {
	return &scheduledTaskRepository{
		etcdClient:  etcdClient,
		redisClient: redisClient,
	}
}

func (s *scheduledTaskRepository) AddTask(ctx context.Context, ttl int64, scheduledTask *biz.ScheduledTask) error {
	leaseResp, err := s.etcdClient.Grant(ctx, ttl)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("%s/%d", taskKey, scheduledTask.Task.TaskId)
	val, err := json.Marshal(scheduledTask)
	if err != nil {
		return err
	}
	_, err = s.etcdClient.Put(ctx, key, string(val), clientv3.WithLease(leaseResp.ID))
	if err != nil {
		return err
	}
	return nil
}

func (s *scheduledTaskRepository) WatchTask(ctx context.Context, execute func(task *biz.ScheduledTask) bool) {
	watchRespChan := s.etcdClient.Watch(context.Background(), taskKey, clientv3.WithPrefix(), clientv3.WithPrevKV())
	for watchResp := range watchRespChan {
		for _, event := range watchResp.Events {
			if event.Type == clientv3.EventTypeDelete {
				logc.Info(ctx, "watch key delete ", zap.String("key", string(event.Kv.Key)))
				value := event.Kv.Value
				var task biz.ScheduledTask
				err := json.Unmarshal(value, &task)
				if err != nil {
					logc.Error(ctx, "unmarshal failed", zap.Error(err))
				} else {
					nx := s.redisClient.SetNX(ctx, fmt.Sprintf("%s:%d", taskLockKey, task.Task.TaskId), 1, 60)
					if nx.Err() != nil {
						logc.Error(ctx, "set lock failed", zap.Error(nx.Err()))
					} else {
						if acquire, err := nx.Result(); err == nil && acquire {
							success := execute(&task)
							if !success {
								s.AddTask(ctx, delaySecondIfFail, &task)
							}
						}
					}
				}
			}
		}
	}
}
