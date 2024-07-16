package service

import (
	"context"
	"github.com/casbin/casbin/v2"
	"github.com/peckfly/gopeck/internal/mods/admin/biz"
)

type CasbinService struct {
	uc *biz.CasbinUsecase
}

func NewCasbinService(uc *biz.CasbinUsecase) *CasbinService {
	return &CasbinService{uc: uc}
}

func (a *CasbinService) GetEnforcer() *casbin.Enforcer {
	if v := a.uc.Enforcer.Load(); v != nil {
		return v.(*casbin.Enforcer)
	}
	return nil
}

func (a *CasbinService) Load(ctx context.Context) error {
	return a.uc.Load(ctx)
}

func (a *CasbinService) Release(ctx context.Context) error {
	return a.uc.Release(ctx)
}
