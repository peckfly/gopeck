package main

import (
	"context"
	"flag"
	"github.com/peckfly/gopeck/internal/conf"
	"github.com/peckfly/gopeck/internal/initialize"
	"github.com/peckfly/gopeck/internal/pkg/common"
	"github.com/peckfly/gopeck/pkg/log"
	"github.com/peckfly/gopeck/pkg/registry"
)

var (
	configFile = flag.String("f", "configs/config-admin.yaml", "the config file")
)

func main() {
	flag.Parse()
	serverConf := conf.ReadConfig(*configFile)
	cleanUp, err := log.Setup(&serverConf.Log)
	log.Must(err)
	etcdClient := initialize.NewEtcdClient(serverConf)
	redisCacheClient := initialize.NewRedisCacheClient(serverConf)
	dbClient := initialize.NewDbClient(serverConf)
	etcdRegistry := registry.NewEtcdRegistry(etcdClient)
	transClient := &common.Trans{DB: dbClient}
	auther, cleanUpAuth, err := initialize.InitAuth(context.Background(), redisCacheClient, &serverConf.Auth)
	log.Must(err)
	srv, err := wireApp(
		serverConf,
		etcdClient,
		dbClient,
		etcdRegistry,
		redisCacheClient,
		transClient,
		auther,
	)
	log.Must(err)
	ctx := context.Background()
	log.Must(srv.Init(ctx))
	log.Must(srv.Run(ctx, []func(){cleanUp, cleanUpAuth}...))
}
