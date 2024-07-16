package biz

import "context"

type (
	Report struct {
		PlanId    uint64 `json:"plan_id"`
		TaskId    uint64 `json:"task_id"`
		Url       string `json:"url"`
		Timestamp int64  `json:"timestamp"`

		TotalNum                   int64            `json:"total_num"`
		TotalResponseContentLength int64            `json:"total_response_content_length"`
		DurationMap                map[int32]int64  `json:"duration_map"`
		StatusMap                  map[int32]int64  `json:"status_map"`
		ErrorMap                   map[string]int64 `json:"error_map"`
		BodyCheckResultMap         map[string]int64 `json:"body_check_result_map"`
		LatencyMap                 map[string]int32 `json:"latency_map"`
	}

	ReporterRepository interface {
		Report(ctx context.Context, detail []*Report) error
	}
)
