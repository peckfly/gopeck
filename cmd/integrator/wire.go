//go:build wireinject
// +build wireinject

package main

import (
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/google/wire"
	"github.com/peckfly/gopeck/internal/conf"
	commondata "github.com/peckfly/gopeck/internal/mods/common/data"
	"github.com/peckfly/gopeck/internal/mods/integrator/biz"
	"github.com/peckfly/gopeck/internal/mods/integrator/data"
	"github.com/peckfly/gopeck/internal/mods/integrator/server"
	"github.com/peckfly/gopeck/internal/mods/integrator/service"
	"github.com/peckfly/gopeck/pkg/cachex"
	"github.com/peckfly/gopeck/pkg/registry"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gorm.io/gorm"
)

func wireApp(*conf.ServerConf, driver.Conn, *clientv3.Client, cachex.Cache, *gorm.DB, registry.Registrar) *server.IntegratorServer {
	panic(wire.Build(
		data.ProviderSet,
		commondata.ProviderSet,
		biz.ProviderSet,
		service.ProviderSet,
		server.ProviderSet,
	))
}
