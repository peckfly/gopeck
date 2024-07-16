package main

import (
	"context"
	"flag"
	"github.com/peckfly/gopeck/internal/conf"
	"github.com/peckfly/gopeck/internal/initialize"
	"github.com/peckfly/gopeck/pkg/log"
	"github.com/peckfly/gopeck/pkg/registry"
)

var (
	configFile = flag.String("f", "configs/config-integrator.yaml", "the config file")
)

func main() {
	flag.Parse()
	serverConf := conf.ReadConfig(*configFile)
	cleanUp, err := log.Setup(&serverConf.Log)
	if err != nil {
		panic(err)
	}
	etcdClient := initialize.NewEtcdClient(serverConf)
	srv := wireApp(serverConf,
		initialize.NewCkClient(serverConf),
		etcdClient,
		initialize.NewRedisCacheClient(serverConf),
		initialize.NewDbClient(serverConf),
		registry.NewEtcdRegistry(etcdClient),
	)
	log.Must(err)
	log.Must(srv.Init())
	log.Must(srv.Run(context.Background(), cleanUp))
}
