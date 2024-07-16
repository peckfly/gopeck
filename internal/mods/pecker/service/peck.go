package service

import (
	"context"
	"github.com/jinzhu/copier"
	v1 "github.com/peckfly/gopeck/api/pecker/v1"
	"github.com/peckfly/gopeck/internal/mods/pecker/biz"
	"github.com/peckfly/gopeck/pkg/log"
	"go.uber.org/zap"
)

type PeckService struct {
	v1.UnimplementedPeckServiceServer
	uc *biz.RequesterUsecase
}

func NewPeckService(PeckUsecase *biz.RequesterUsecase) *PeckService {
	return &PeckService{
		uc: PeckUsecase,
	}
}

func (r PeckService) Peck(ctx context.Context, in *v1.PeckRequest) (*v1.PeckReply, error) {
	var peck biz.Requester
	err := copier.Copy(&peck, in)
	if err != nil {
		log.Error("Peck failed to copy request", zap.Error(err))
		return nil, err
	}
	err = r.uc.Request(ctx, &peck)
	if err != nil {
		log.Error("Peck failed", zap.Error(err))
		return nil, err
	}
	// todo response code
	return &v1.PeckReply{}, nil
}

func (r PeckService) Stop(ctx context.Context, in *v1.StopRequest) (*v1.StopReply, error) {
	err := r.uc.Stop(ctx, in.PlanId, in.TaskId)
	if err != nil {
		log.Error("stop failed", zap.Error(err))
		return nil, err
	}
	return &v1.StopReply{}, nil
}
