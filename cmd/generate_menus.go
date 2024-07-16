package main

import (
	"context"
	"encoding/json"
	"flag"
	"github.com/peckfly/gopeck/internal/conf"
	"github.com/peckfly/gopeck/internal/initialize"
	"github.com/peckfly/gopeck/internal/mods/admin/biz"
	"github.com/peckfly/gopeck/internal/mods/admin/data"
	"github.com/peckfly/gopeck/internal/pkg/common"
	"os"
)

var (
	configFileGenMenus = flag.String("f", "configs/config-admin.yaml", "the config file")
)

func main() {
	flag.Parse()
	serverConf := conf.ReadConfig(*configFileGenMenus)
	dbClient := initialize.NewDbClient(serverConf)
	redisClient := initialize.NewRedisCacheClient(serverConf)
	transClient := &common.Trans{DB: dbClient}
	f, err := os.ReadFile("configs/menu.json")
	if err != nil {
		panic(err)
	}
	var menus biz.Menus
	if err := json.Unmarshal(f, &menus); err != nil {
		panic(err)
	}
	a := biz.NewMenuUsecase(redisClient, transClient, data.NewMenuRepository(dbClient), data.NewMenuResourceRepository(dbClient), data.NewRoleMenuRepository(dbClient), serverConf)
	err = transClient.Exec(context.Background(), func(ctx context.Context) error {
		return a.CreateInBatchByParent(ctx, menus, nil)
	})
	if err != nil {
		panic(err)
	}
}
