package biz

import (
	"context"
	v1 "github.com/peckfly/gopeck/api/pecker/v1"
	"github.com/peckfly/gopeck/internal/mods/common/repo"
	"github.com/peckfly/gopeck/internal/pkg/enums"
	"github.com/peckfly/gopeck/pkg/log/logc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	grpcinsecure "google.golang.org/grpc/credentials/insecure"
	"strconv"
	"strings"
	"time"
)

// StopStress stops stress for a given plan ID by stopping all associated tasks.
//
// ctx context.Context, b *Stop
// error
func (s *StressUsecase) StopStress(ctx context.Context, b *Stop) error {
	planId, err := strconv.ParseUint(b.PlanId, 10, 64)
	if err != nil {
		return err
	}
	tasks, err := s.recordRepository.FindTaskListByPlanId(ctx, planId)
	if err != nil {
		return err
	}
	if len(tasks) == 0 {
		return nil
	}
	for _, t := range tasks {
		s.Stop(ctx, planId, t)
	}
	return nil
}

// Stop stops the task execution on the specified nodes.
//
// Parameters:
// - ctx: the context for the request.
// - planId: the ID of the plan to stop.
// - task: the task record containing information about the nodes.
func (s *StressUsecase) Stop(ctx context.Context, planId uint64, task *repo.TaskRecord) {
	nodeAddrs := strings.Split(task.Nodes, ",")
	for _, addr := range nodeAddrs {
		conn, err := grpc.DialContext(ctx, strings.ReplaceAll(addr, "grpc://", ""),
			grpc.WithTransportCredentials(grpcinsecure.NewCredentials()))
		if err != nil {
			logc.Error(ctx, "dial error", zap.String("addr", addr), zap.Error(err))
			continue
		}
		PeckServiceClient := v1.NewPeckServiceClient(conn)
		_, err = PeckServiceClient.Stop(ctx, &v1.StopRequest{PlanId: planId, TaskId: task.TaskId})
		if err != nil {
			logc.Error(ctx, "stop task error", zap.Error(err))
			continue
		}
		err = conn.Close()
		if err != nil {
			logc.Error(ctx, "close conn error", zap.Error(err))
		}
	}
	err := s.recordRepository.UpdatePlanById(ctx, planId, &repo.PlanRecord{Status: int(enums.STOP), UpdateTime: time.Now().Unix()})
	if err != nil {
		logc.Error(ctx, "update plan record error", zap.Error(err))
	}
}
