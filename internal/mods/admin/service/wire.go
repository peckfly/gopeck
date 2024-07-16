package service

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewCasbinService,
	NewLoginService,
	NewMenuService,
	NewRoleService,
	NewStressService,
	NewUserService,
	NewNodeService,
)
