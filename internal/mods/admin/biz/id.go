package biz

import (
	"github.com/sony/sonyflake"
	"time"
)

var planSf, taskSf *sonyflake.Sonyflake

func init() {
	planSf = sonyflake.NewSonyflake(sonyflake.Settings{
		StartTime: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	taskSf = sonyflake.NewSonyflake(sonyflake.Settings{
		StartTime: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
	})
}

func generatePlanId() (uint64, error) {
	return planSf.NextID()
}

func generateTaskId() (uint64, error) {
	return taskSf.NextID()
}
