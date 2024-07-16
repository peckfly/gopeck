package service

import (
	"context"
	"github.com/jinzhu/copier"
	v1 "github.com/peckfly/gopeck/api/integrator/v1"
	"github.com/peckfly/gopeck/internal/mods/integrator/biz"
	"github.com/peckfly/gopeck/pkg/log"
	"go.uber.org/zap"
)

type IntegrateService struct {
	v1.UnimplementedIntegrateServiceServer
	uc *biz.IntegratorUsecase
}

func NewIntegrateService(uc *biz.IntegratorUsecase) *IntegrateService {
	return &IntegrateService{uc: uc}
}

func (s *IntegrateService) Integrate(ctx context.Context, req *v1.IntegrateRequest) (*v1.IntegrateReply, error) {
	var integrate biz.Integrate
	err := copier.Copy(&integrate, req)
	if err != nil {
		log.Error("Integrate failed to copy request", zap.Error(err))
		return nil, err
	}
	err = s.uc.IntegrateReport(ctx, &integrate)
	if err != nil {
		log.Error("Integrate failed", zap.Error(err))
		return nil, err
	}
	return &v1.IntegrateReply{}, nil
}
