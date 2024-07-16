package repo

import (
	"context"
	"time"
)

type (
	QueRepository interface {
		AggregatePush(ctx context.Context, taskId uint64, result *Aggregate) error

		AggregatePop(ctx context.Context, taskId uint64) (*Aggregate, error)

		AggregateClear(ctx context.Context, taskId uint64) error

		RatePush(ctx context.Context, taskId uint64, result *Aggregate) error

		RatePop(ctx context.Context, taskId uint64) (*Aggregate, error)

		RateClear(ctx context.Context, taskId uint64) error

		BatchSetTaskNodeCount(ctx context.Context, counts map[uint64]int) error

		GetTaskNodeCount(ctx context.Context, taskId uint64) (int, error)
	}

	Result struct {
		Interval              int           `json:"interval"`
		Err                   string        `json:"err"`
		StatusCode            int           `json:"status_code"`
		Duration              time.Duration `json:"duration"`
		ResponseContentLength int64         `json:"response_content_length"`
		TimeStamp             int64         `json:"timestamp"`
		Stop                  bool          `json:"stop"`
		BodyCheckResult       string        `json:"body_check_result"`
	}

	Aggregate struct {
		Interval                   int              `json:"interval"`
		PlanId                     uint64           `json:"plan_id"`
		TaskId                     uint64           `json:"task_id"`
		Timestamp                  int64            `json:"timestamp"`
		TotalNum                   int64            `json:"total_num"`
		TotalResponseContentLength int64            `json:"total_response_content_length"`
		DurationMap                map[int32]int64  `json:"duration_map"`
		StatusMap                  map[int32]int64  `json:"status_map"`
		ErrorMap                   map[string]int64 `json:"error_map"`
		BodyCheckResultMap         map[string]int64 `json:"body_check_result_map"`
		Stop                       bool             `json:"stop"`
	}
)

func NewAggeRate() *Aggregate {
	return &Aggregate{
		TotalNum:                   int64(0),
		TotalResponseContentLength: int64(0),
		DurationMap:                make(map[int32]int64),
		StatusMap:                  make(map[int32]int64),
		ErrorMap:                   make(map[string]int64),
		BodyCheckResultMap:         make(map[string]int64),
	}
}
