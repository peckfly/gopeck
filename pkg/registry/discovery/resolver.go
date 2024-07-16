package discovery

import (
	"context"
	"errors"
	"github.com/peckfly/gopeck/pkg/log"
	"github.com/peckfly/gopeck/pkg/registry"
	"github.com/peckfly/gopeck/pkg/registry/endpoint"
	"go.uber.org/zap"
	"time"

	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"
)

type discoveryResolver struct {
	w  registry.Watcher
	cc resolver.ClientConn

	ctx    context.Context
	cancel context.CancelFunc

	insecure bool
}

func (r *discoveryResolver) watch() {
	for {
		select {
		case <-r.ctx.Done():
			return
		default:
		}
		ins, err := r.w.Next()
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			log.Error("[resolver] Failed to watch discovery endpoint", zap.Error(err))
			time.Sleep(time.Second)
			continue
		}
		r.update(ins)
	}
}

func (r *discoveryResolver) update(ins []*registry.ServiceInstance) {
	var (
		endpoints = make(map[string]struct{})
		filtered  = make([]*registry.ServiceInstance, 0, len(ins))
	)
	for _, in := range ins {
		ept, err := endpoint.ParseEndpoint([]string{in.Addr}, endpoint.Scheme("grpc", !r.insecure))
		if err != nil {
			log.Error("[resolver] Failed to parse discovery endpoint", zap.Error(err))
			continue
		}
		if len(ept) <= 0 {
			continue
		}
		// filter redundant endpoints
		if _, ok := endpoints[ept]; ok {
			continue
		}
		filtered = append(filtered, in)
	}

	addrs := make([]resolver.Address, 0, len(filtered))
	for _, in := range filtered {
		ept, _ := endpoint.ParseEndpoint([]string{in.Addr}, endpoint.Scheme("grpc", !r.insecure))
		endpoints[ept] = struct{}{}
		addr := resolver.Address{
			ServerName: in.Name,
			Attributes: parseAttributes(in.Metadata).WithValue("rawServiceInstance", in),
			Addr:       ept,
		}
		addrs = append(addrs, addr)
	}
	if len(addrs) == 0 {
		log.Warn("[resolver] Zero endpoint found, refused to write", zap.Any("instances", ins))
		return
	}
	err := r.cc.UpdateState(resolver.State{Addresses: addrs})
	if err != nil {
		log.Error("[resolver] failed to update state", zap.Error(err))
	}
}

func (r *discoveryResolver) Close() {
	r.cancel()
	err := r.w.Stop()
	if err != nil {
		log.Error("[resolver] failed to watch top", zap.Error(err))
	}
}

func (r *discoveryResolver) ResolveNow(_ resolver.ResolveNowOptions) {}

func parseAttributes(md map[string]string) (a *attributes.Attributes) {
	for k, v := range md {
		a = a.WithValue(k, v)
	}
	return a
}
