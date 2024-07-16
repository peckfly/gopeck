package biz

import (
	"github.com/peckfly/gopeck/internal/mods/common/repo"
	"github.com/peckfly/gopeck/internal/pkg/common"
	"github.com/peckfly/gopeck/pkg/registry"
	"google.golang.org/grpc"
	"strings"
)

// stress
type (
	Plan struct {
		PlanId           uint64 `json:"-"`
		PlanName         string `json:"plan_name"`
		StartTime        int64  `json:"start_time"` // todo if StartTime > 0 then use scheduleTask
		StressTime       int    `json:"stress_time" binding:"required,min=1"`
		StressType       int    `json:"stress_type" binding:"required,min=1,max=2"`
		StressMode       int    `json:"stress_mode" binding:"required,min=1,max=2"`
		UserId           string `json:"user_id"`
		Tasks            []Task `json:"tasks"`
		StepIntervalTime int    `json:"step_interval_time"`
		IntervalLen      int    `json:"-"`
	}
	Task struct {
		PlanId           uint64 `json:"-"`
		TaskId           uint64 `json:"-"`
		StressTime       int    `json:"-"`
		StressType       int    `json:"-"`
		StressMode       int    `json:"-"`
		StepIntervalTime int    `json:"-"`
		TaskName         string `json:"task_name" binding:"required"`
		Num              int    `json:"num" binding:"required,min=1"`
		MaxNum           int    `json:"max_num"`
		StepNum          int    `json:"step_num"`
		MaxConnections   int    `json:"max_connections" binding:"min=1"`
		Url              string `json:"url" binding:"required"`
		Method           string `json:"method" binding:"required,oneof=GET POST PUT DELETE"`
		Timeout          int    `json:"timeout" binding:"min=1,max=5000"`

		QueryEntry          []Entry           `json:"query"`
		HeaderEntry         []Entry           `json:"header"`
		Query               map[string]string `json:"-"`
		Headers             map[string]string `json:"-"`
		Body                string            `json:"body"`
		DynamicParamScript  string            `json:"dynamic_param_script"`
		ResponseCheckScript string            `json:"response_check_script"`

		DisableCompression bool `json:"-"`
		DisableKeepAlive   bool `json:"-"`
		DisableRedirects   bool `json:"-"`
		H2                 bool `json:"-"`

		Options []string `json:"options"`

		Proxy string `json:"proxy"`

		DynamicParams []DynamicParam `json:"-"`
		nodes         []*BindNode
	}
	Entry struct {
		EntryKey   string `json:"entry_key"`
		EntryValue string `json:"entry_value"`
	}

	Stop struct {
		PlanId string `json:"plan_id"`
		UserId string `json:"user_id"`
	}
	BindNode struct {
		num      int
		Nums     []int32
		nodeInfo *repo.Node
	}
	BindConn struct {
		num      int
		Nums     []int32
		Addr     string
		grpcConn *grpc.ClientConn
	}
	NodeInstanceCost struct {
		instance         *registry.ServiceInstance
		RpsQuota         int
		GoroutineQuota   int
		RpsCost          int
		GoroutineCost    int
		RunningTaskCount int
		UnExist          bool
	}
	DynamicParam struct {
		Headers map[string]string
		Query   map[string]string
		Body    string
	}
	Restart struct {
		PlanId uint64
		UserId string
	}

	PlanRecordQuery struct {
		common.PaginationParam
		PlanId    uint64 `form:"plan_id"`
		PlanName  string `form:"plan_name"`
		StartTime int64  `form:"start_time"`
		EndTime   int64  `form:"end_time"`
	}

	TaskRecordQuery struct {
		PlanId string `form:"plan_id"`
	}
)

// record
type (
	PlanQueryResult struct {
		Data       []*PlanRecordResultItem
		PageResult *common.PaginationResult
	}
	TaskResult struct {
		Data []*TaskResultItem `json:"data"`
	}

	PlanRecordResultItem struct {
		PlanId             string  `json:"plan_id"`
		UserId             string  `json:"user_id"`
		StartTime          int64   `json:"start_time"`
		StressTime         int     `json:"stress_time"`
		StressType         int     `json:"stress_type"`
		StressMode         int     `json:"stress_mode"`
		Status             int     `json:"status"`
		StressProgress     float64 `json:"stress_progress"`
		OverviewMetricsUrl string  `json:"overview_metrics_url"`
		PlanName           string  `json:"plan_name"`
		CreateTime         int64   `json:"create_time"`
		UpdateTime         int64   `json:"update_time"`
	}

	TaskResultItem struct {
		TaskId              string    `json:"task_id"`
		TaskName            string    `json:"task_name"`
		TaskStatus          int       `json:"task_status"`
		PlanId              string    `json:"plan_id"`
		Url                 string    `json:"url"`
		StressType          int       `json:"stress_type"`
		StressMode          int       `json:"stress_mode"`
		Num                 int       `json:"num" binding:"required,min=1"`
		MaxNum              int       `json:"max_num"`
		StepNum             int       `json:"step_num"`
		Timeout             int       `json:"timeout" `
		MaxConnections      int       `json:"max_connections"`
		ProtocolType        int       `json:"protocol_type"`
		Method              string    `json:"method"`
		QueryEntry          []Entry   `json:"query"`
		HeaderEntry         []Entry   `json:"header"`
		Body                *JsonBody `json:"body"`
		DynamicParamScript  string    `json:"dynamic_param_script"`
		ResponseCheckScript string    `json:"response_check_script"`

		DisableCompression int8   `json:"disable_compression"`
		DisableKeepAlive   int8   `json:"disable_keep_alive"`
		DisableRedirects   int8   `json:"disable_redirects"`
		H2                 int8   `json:"h_2"`
		Proxy              string `json:"proxy"`

		MaxBodySize int64 `json:"max_body_size"`

		Nodes string `json:"nodes"`

		MetricsUrl string `json:"metrics_url"`

		Reports []*Summary `json:"reports"`

		CreateTime int64 `json:"create_time"`
		UpdateTime int64 `json:"update_time"`
	}

	JsonBody struct {
		Json map[string]any `json:"json"`
	}

	Summary struct {
		Num    int
		TaskId uint64
		Rps    float64

		AvgTotal      int64
		Fastest       int64
		Slowest       int64
		Average       float64
		ErrorCount    int64
		ErrorRate     string
		SizeTotal     int64
		NumRes        int64
		TotalCostTime float64

		Lat90  int `json:"lat_90"`
		Lat95  int `json:"lat_95"`
		Lat99  int `json:"lat_99"`
		Lat999 int `json:"lat_999"`

		TimeBuckets         []int64 `json:"-"`
		StatusCodeDist      map[int]int64
		Histogram           []Bucket
		LatencyDistribution []LatencyDistribution
		ErrorDist           map[string]int64
		BodyCheckResultMap  map[string]int64
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
)

// node
type (
	NodeQuery struct {
		Addr string `form:"addr"`
	}

	NodeListResult struct {
		Data []*NodeResultItem `json:"data"`
	}

	NodeResultItem struct {
		Addr             string `json:"addr"`
		RpsQuota         int    `json:"rps_quota"`
		GoroutineQuota   int    `json:"goroutine_quota"`
		RpsCost          int    `json:"rps_cost"`
		GoroutineCost    int    `json:"goroutine_cost"`
		RunningTaskCount int    `json:"running_task_count"`
	}
	NodeQueryResult struct {
		Data []*repo.NodeState `json:"data"`
	}

	UpdateNodeForm struct {
		Addr           string `json:"addr"`
		RpsQuota       int    `json:"rps_quota"`
		GoroutineQuota int    `json:"goroutine_quota"`
	}
)

// login
type (
	CaptchaResult struct {
		CaptchaID string `json:"captcha_id"` // Captcha ID
	}

	LoginResult struct {
		AccessToken string `json:"access_token"` // Access token (JWT)
		TokenType   string `json:"token_type"`   // Token type (Usage: Authorization=${token_type} ${access_token})
		ExpiresAt   int64  `json:"expires_at"`   // Expired time (Unit: second)
	}
	LoginForm struct {
		Username    string `json:"username" binding:"required"`     // Login name
		Password    string `json:"password" binding:"required"`     // Login password (md5 hash)
		CaptchaID   string `json:"captcha_id" binding:"required"`   // Captcha verify id
		CaptchaCode string `json:"captcha_code" binding:"required"` // Captcha verify code
	}
	UpdateLoginPassword struct {
		OldPassword string `json:"old_password" binding:"required"` // Old password (md5 hash)
		NewPassword string `json:"new_password" binding:"required"` // New password (md5 hash)
	}
	UpdateCurrentUser struct {
		Name   string `json:"name" binding:"required,max=64"` // Name of user
		Phone  string `json:"phone" binding:"max=32"`         // Phone number of user
		Email  string `json:"email" binding:"max=128"`        // Email of user
		Remark string `json:"remark" binding:"max=1024"`      // Remark of user
	}
)

func (a *LoginForm) Trim() *LoginForm {
	a.Username = strings.TrimSpace(a.Username)
	a.CaptchaCode = strings.TrimSpace(a.CaptchaCode)
	return a
}
