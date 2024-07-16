package biz

import (
	"context"
	"encoding/json"
	"github.com/jinzhu/copier"
	"github.com/peckfly/gopeck/pkg/netx"
)

// MaxPullTaskCount max pull task count by planId, generally 200 is enough
const MaxPullTaskCount = 200

func (s *StressUsecase) RestartStress(ctx context.Context, restart *Restart) error {
	planRecord, err := s.recordRepository.FindPlanRecordByPlanId(ctx, restart.PlanId)
	if err != nil {
		return err
	}
	taskRecords, err := s.recordRepository.FindTaskListByPlanIdWithSize(ctx, restart.PlanId, MaxPullTaskCount)
	if err != nil {
		return err
	}
	var tasks []Task
	for _, taskRecord := range taskRecords {
		var task Task
		err = copier.Copy(&task, taskRecord)
		if err != nil {
			return err
		}
		task.Query = netx.ParseQuery(taskRecord.Query)
		headers := make(map[string]string)
		err = json.Unmarshal([]byte(taskRecord.Header), &headers)
		if err != nil {
			return err
		}
		task.Headers = headers
		task.DisableRedirects = taskRecord.DisableRedirects == 1
		task.DisableKeepAlive = taskRecord.DisableKeepAlive == 1
		task.DisableCompression = taskRecord.DisableCompression == 1
		task.H2 = taskRecord.H2 == 1
		tasks = append(tasks, task)
	}
	// todo startTime design
	return s.StartStress(ctx, &Plan{
		StressTime: planRecord.StressTime,
		UserId:     restart.UserId,
		PlanName:   planRecord.PlanName,
		Tasks:      tasks,
	})
}
