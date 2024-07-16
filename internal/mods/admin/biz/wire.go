package biz

import (
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	NewCasbinUsecase,
	NewLoginUsecase,
	NewMenuUsecase,
	NewRoleUsecase,
	NewStressUsecase,
	NewUserUsecase,
	NewNodeUsecase,
)
