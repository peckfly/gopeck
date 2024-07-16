package data

import (
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	NewMenuRepository,
	NewMenuResourceRepository,
	NewRoleRepository,
	NewRoleMenuRepository,
	NewScheduledTaskRepository,
	NewUserRepository,
	NewUserRoleRepository,
)
