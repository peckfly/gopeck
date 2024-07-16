package repo

import (
	"context"
	"github.com/peckfly/gopeck/internal/pkg/common"
)

type (
	PlanRecord struct {
		PlanId           uint64 `gorm:"column:plan_id" json:"plan_id"`
		UserId           string `gorm:"column:user_id" json:"user_id"`
		StartTime        int64  `gorm:"column:start_time" json:"start_time"`
		StressTime       int    `gorm:"column:stress_time" json:"stress_time"`
		StressType       int    `gorm:"column:stress_type" json:"stress_type"`
		Status           int    `gorm:"column:status" json:"status"`
		StressMode       int    `gorm:"column:stress_mode" json:"stress_mode"`
		PlanName         string `gorm:"column:plan_name" json:"plan_name"`
		StepIntervalTime int    `gorm:"column:step_interval_time" json:"step_interval_time"`
		IntervalLen      int    `gorm:"column:interval_len" json:"interval_len"`
		CreateTime       int64  `gorm:"column:create_time" json:"create_time"`
		UpdateTime       int64  `gorm:"column:update_time" json:"update_time"`
	}

	TaskRecord struct {
		TaskId              uint64 `gorm:"column:task_id" json:"task_id"`
		TaskName            string `gorm:"column:task_name" json:"task_name"`
		TaskStatus          int    `gorm:"column:task_status" json:"task_status"`
		PlanId              uint64 `gorm:"column:plan_id" json:"plan_id"`
		Url                 string `gorm:"column:url" json:"url"`
		StressType          int    `gorm:"column:stress_type" json:"stress_type"`
		StressMode          int    `gorm:"column:stress_mode" json:"stress_mode"`
		Num                 int    `gorm:"column:num" json:"num" binding:"required,min=1"`
		MaxNum              int    `gorm:"column:max_num" json:"max_num"`
		StepNum             int    `gorm:"column:step_num" json:"step_num"`
		Timeout             int    `gorm:"column:timeout" json:"timeout" `
		MaxConnections      int    `gorm:"column:max_connections" json:"max_connections"`
		ProtocolType        int    `gorm:"column:protocol_type" json:"protocol_type"`
		Method              string `gorm:"column:method" json:"method"`
		Query               string `gorm:"column:query" json:"query"`
		Header              string `gorm:"column:header" json:"header"`
		Body                string `gorm:"column:body" json:"body"`
		DynamicParamScript  string `gorm:"column:dynamic_param_script" json:"dynamic_param_script"`
		ResponseCheckScript string `gorm:"column:response_check_script" json:"response_check_script"`

		DisableCompression int8   `gorm:"column:disable_compression" json:"disable_compression"`
		DisableKeepAlive   int8   `gorm:"column:disable_keep_alive" json:"disable_keep_alive"`
		DisableRedirects   int8   `gorm:"column:disable_redirects" json:"disable_redirects"`
		H2                 int8   `gorm:"column:h_2" json:"h_2"`
		Proxy              string `gorm:"column:proxy" json:"proxy"`

		MaxBodySize int64 `gorm:"column:max_body_size" json:"max_body_size"`

		// stress node list
		Nodes string `gorm:"column:nodes" json:"nodes"`

		// StatExt
		StatExt string `gorm:"column:stat_ext" json:"stat_ext"`

		CreateTime int64 `gorm:"column:create_time" json:"create_time"`
		UpdateTime int64 `gorm:"column:update_time" json:"update_time"`
	}

	StatExt struct {
		ErrorDist           map[string]int64
		StatusCodeDist      map[int]int64
		Histogram           []Bucket
		LatencyDistribution []LatencyDistribution
	}

	LatencyDistribution struct {
		Percentage    float64
		Latency       int
		PercentageStr string
		LatencyMs     string
	}

	Bucket struct {
		Mark      int
		Count     int64
		Frequency float64
	}

	RecordRepository interface {
		CreatePlan(context.Context, *PlanRecord) error
		BatchCreateTasks(ctx context.Context, records []TaskRecord) error
		UpdateTaskById(ctx context.Context, taskId uint64, record *TaskRecord) error
		UpdatePlanById(ctx context.Context, planId uint64, record *PlanRecord) error
		FindTaskListByPlanId(ctx context.Context, planId uint64) ([]*TaskRecord, error)
		QueryUserRecordsByUserId(ctx context.Context, userId string, planId uint64, planName string, startTime, endTime int64, pp common.PaginationParam) ([]*PlanRecord, *common.PaginationResult, error)
		QueryTaskRecordsByPlanId(ctx context.Context, planId uint64) (records []*TaskRecord, err error)
		FindPlanRecordByPlanId(ctx context.Context, planId uint64) (*PlanRecord, error)
		FindTaskListByPlanIdWithSize(ctx context.Context, planId uint64, count int) (records []*TaskRecord, err error)
	}
)
