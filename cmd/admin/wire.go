//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/peckfly/gopeck/internal/conf"
	"github.com/peckfly/gopeck/internal/mods/admin/biz"
	"github.com/peckfly/gopeck/internal/mods/admin/data"
	"github.com/peckfly/gopeck/internal/mods/admin/server"
	"github.com/peckfly/gopeck/internal/mods/admin/service"
	commondata "github.com/peckfly/gopeck/internal/mods/common/data"
	"github.com/peckfly/gopeck/internal/pkg/common"
	"github.com/peckfly/gopeck/internal/pkg/jwtx"
	"github.com/peckfly/gopeck/pkg/cachex"
	"github.com/peckfly/gopeck/pkg/registry"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gorm.io/gorm"
)

func wireApp(
	*conf.ServerConf,
	*clientv3.Client,
	*gorm.DB,
	registry.Discovery,
	cachex.Cache,
	*common.Trans,
	jwtx.Auther,
) (*server.AdminServer, error) {
	panic(wire.Build(
		data.ProviderSet,
		commondata.ProviderSet,
		biz.ProviderSet,
		service.ProviderSet,
		server.ProviderSet,
	))
}
