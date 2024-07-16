package biz

type (
	Integrate struct {
		PlanId           uint64
		Tasks            []Task
		StressTime       int32
		StressType       int32
		StressMode       int32
		StepIntervalTime int32
		StartTime        int64
		IntervalLen      int32
		UserId           int64
	}

	Task struct {
		TaskId               uint64
		Url                  string
		RequestContentLength int
	}

	Summary struct {
		TaskId uint64
		Rps    float64

		AvgTotal      int64
		Fastest       int64
		Slowest       int64
		Average       float64
		ErrorCount    int64
		SizeTotal     int64
		NumRes        int64
		TotalCostTime float64

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
