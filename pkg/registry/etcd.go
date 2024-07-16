package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/peckfly/gopeck/internal/pkg/consts"
	"math/rand"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// Option is etcd EtcdRegistry option.
type Option func(o *options)

type options struct {
	ctx       context.Context
	namespace string
	ttl       time.Duration
	maxRetry  int
}

// EtcdRegistry is etcd EtcdRegistry.
type EtcdRegistry struct {
	opts   *options
	client *clientv3.Client
	kv     clientv3.KV
	lease  clientv3.Lease
	/*
		ctxMap is used to store the context cancel function of each service instance.
		When the service instance is Unregistered, the corresponding context cancel function is called to stop the heartbeat.
	*/
	ctxMap map[*ServiceInstance]context.CancelFunc
}

// NewEtcdRegistry creates etcd EtcdRegistry
func NewEtcdRegistry(client *clientv3.Client, opts ...Option) (r *EtcdRegistry) {
	op := &options{
		ctx:       context.Background(),
		namespace: "/grpc-mirco/" + consts.AppName,
		ttl:       time.Second * 10,
		maxRetry:  5,
	}
	for _, o := range opts {
		o(op)
	}
	return &EtcdRegistry{
		opts:   op,
		client: client,
		kv:     clientv3.NewKV(client),
		ctxMap: make(map[*ServiceInstance]context.CancelFunc),
	}
}

// WithContext with EtcdRegistry context.
func WithContext(ctx context.Context) Option {
	return func(o *options) { o.ctx = ctx }
}

// WithNamespace with EtcdRegistry namespace.
func WithNamespace(ns string) Option {
	return func(o *options) { o.namespace = ns }
}

// WithRegisterTTL with register ttl.
func WithRegisterTTL(ttl time.Duration) Option {
	return func(o *options) { o.ttl = ttl }
}

// WithMaxRetry with max retry.
func WithMaxRetry(num int) Option {
	return func(o *options) { o.maxRetry = num }
}

// Register the registration.
func (r *EtcdRegistry) Register(ctx context.Context, service *ServiceInstance) error {
	key := fmt.Sprintf("%s/%s/%s", r.opts.namespace, service.Name, service.Addr)
	value, err := json.Marshal(service)
	if err != nil {
		return err
	}
	if r.lease != nil {
		r.lease.Close()
	}
	r.lease = clientv3.NewLease(r.client)
	leaseID, err := r.registerWithKV(ctx, key, string(value))
	if err != nil {
		return err
	}

	hCtx, cancel := context.WithCancel(r.opts.ctx)
	r.ctxMap[service] = cancel
	go r.heartBeat(hCtx, leaseID, key, string(value))
	return nil
}

// Unregister the registration.
func (r *EtcdRegistry) Unregister(ctx context.Context, service *ServiceInstance) error {
	defer func() {
		if r.lease != nil {
			r.lease.Close()
		}
	}()
	// cancel heartbeat
	if cancel, ok := r.ctxMap[service]; ok {
		cancel()
		delete(r.ctxMap, service)
	}
	key := fmt.Sprintf("%s/%s/%s", r.opts.namespace, service.Name, service.Addr)
	_, err := r.client.Delete(ctx, key)
	return err
}

// GetService return the service instances in memory according to the service name.
func (r *EtcdRegistry) GetService(ctx context.Context, name string) ([]*ServiceInstance, error) {
	key := fmt.Sprintf("%s/%s", r.opts.namespace, name)
	resp, err := r.kv.Get(ctx, key, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	items := make([]*ServiceInstance, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		si, err := unmarshal(kv.Value)
		if err != nil {
			return nil, err
		}
		if si.Name != name {
			continue
		}
		items = append(items, si)
	}
	return items, nil
}

// Watch creates a watcher according to the service name.
func (r *EtcdRegistry) Watch(ctx context.Context, name string) (Watcher, error) {
	key := fmt.Sprintf("%s/%s", r.opts.namespace, name)
	return newWatcher(ctx, key, name, r.client)
}

// registerWithKV create a new lease, return current leaseID
func (r *EtcdRegistry) registerWithKV(ctx context.Context, key string, value string) (clientv3.LeaseID, error) {
	grant, err := r.lease.Grant(ctx, int64(r.opts.ttl.Seconds()))
	if err != nil {
		return 0, err
	}
	_, err = r.client.Put(ctx, key, value, clientv3.WithLease(grant.ID))
	if err != nil {
		return 0, err
	}
	return grant.ID, nil
}

func (r *EtcdRegistry) heartBeat(ctx context.Context, leaseID clientv3.LeaseID, key string, value string) {
	curLeaseID := leaseID
	kac, err := r.client.KeepAlive(ctx, leaseID)
	if err != nil {
		curLeaseID = 0
	}
	rand.NewSource(time.Now().Unix())

	for {
		if curLeaseID == 0 {
			// try to registerWithKV
			var retreat []int
			for retryCnt := 0; retryCnt < r.opts.maxRetry; retryCnt++ {
				if ctx.Err() != nil {
					return
				}
				// prevent infinite blocking
				idChan := make(chan clientv3.LeaseID, 1)
				errChan := make(chan error, 1)
				cancelCtx, cancel := context.WithCancel(ctx)
				go func() {
					defer cancel()
					id, registerErr := r.registerWithKV(cancelCtx, key, value)
					if registerErr != nil {
						errChan <- registerErr
					} else {
						idChan <- id
					}
				}()

				select {
				case <-time.After(3 * time.Second):
					cancel()
					continue
				case <-errChan:
					continue
				case curLeaseID = <-idChan:
				}

				kac, err = r.client.KeepAlive(ctx, curLeaseID)
				if err == nil {
					break
				}
				retreat = append(retreat, 1<<retryCnt)
				time.Sleep(time.Duration(retreat[rand.Intn(len(retreat))]) * time.Second)
			}
			if _, ok := <-kac; !ok {
				// retry failed
				return
			}
		}

		select {
		case _, ok := <-kac:
			if !ok {
				if ctx.Err() != nil {
					// channel closed due to context cancel
					return
				}
				// need to retry registration
				curLeaseID = 0
				continue
			}
		case <-r.opts.ctx.Done():
			return
		}
	}
}

func unmarshal(data []byte) (si *ServiceInstance, err error) {
	err = json.Unmarshal(data, &si)
	return
}
