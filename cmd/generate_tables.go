package main

import (
	"flag"
	"github.com/peckfly/gopeck/internal/conf"
	"github.com/peckfly/gopeck/internal/initialize"
	"github.com/peckfly/gopeck/internal/mods/admin/biz"
	"github.com/peckfly/gopeck/internal/mods/common/repo"
)

var (
	configFileGenTables = flag.String("f", "configs/config-admin.yaml", "the config file")
)

func main() {
	flag.Parse()
	serverConf := conf.ReadConfig(*configFileGenTables)
	dbClient := initialize.NewDbClient(serverConf)
	err := dbClient.AutoMigrate(
		new(biz.Menu),
		new(biz.MenuResource),
		new(biz.Role),
		new(biz.RoleMenu),
		new(biz.User),
		new(biz.UserRole),
		new(repo.PlanRecord),
		new(repo.TaskRecord),
	)
	if err != nil {
		panic(err)
	}
}
