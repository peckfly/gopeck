// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/peckfly/gopeck/internal/conf"
	"github.com/peckfly/gopeck/internal/mods/common/data"
	"github.com/peckfly/gopeck/internal/mods/pecker/biz"
	"github.com/peckfly/gopeck/internal/mods/pecker/server"
	"github.com/peckfly/gopeck/internal/mods/pecker/service"
	"github.com/peckfly/gopeck/pkg/cachex"
	"github.com/peckfly/gopeck/pkg/registry"
	"go.etcd.io/etcd/client/v3"
)

// Injectors from wire.go:

func wireApp(serverConf *conf.ServerConf, client *clientv3.Client, cache cachex.Cache, registrar registry.Registrar, discovery registry.Discovery) *server.PeckerServer {
	queRepository := data.NewQueRepository(cache)
	nodeRepository := data.NewNodeRepository(client, cache)
	requesterUsecase := biz.NewRequesterUsecase(serverConf, queRepository, nodeRepository, discovery)
	peckService := service.NewPeckService(requesterUsecase)
	peckerServer := server.NewPeckerServer(serverConf, peckService, registrar, nodeRepository)
	return peckerServer
}
