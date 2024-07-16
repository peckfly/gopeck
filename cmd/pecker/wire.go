//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/peckfly/gopeck/internal/conf"
	commondata "github.com/peckfly/gopeck/internal/mods/common/data"
	"github.com/peckfly/gopeck/internal/mods/pecker/biz"
	"github.com/peckfly/gopeck/internal/mods/pecker/server"
	"github.com/peckfly/gopeck/internal/mods/pecker/service"
	"github.com/peckfly/gopeck/pkg/cachex"
	"github.com/peckfly/gopeck/pkg/registry"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func wireApp(*conf.ServerConf, *clientv3.Client, cachex.Cache, registry.Registrar, registry.Discovery) *server.PeckerServer {
	panic(wire.Build(
		biz.ProviderSet,
		commondata.ProviderSet,
		service.ProviderSet,
		server.ProviderSet,
	))
}
