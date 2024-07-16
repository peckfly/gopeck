package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/panjf2000/ants/v2"
	"github.com/peckfly/gopeck/internal/mods/common/repo"
	"github.com/peckfly/gopeck/internal/pkg/enums"
	"github.com/peckfly/gopeck/pkg/log/logc"
	"go.uber.org/zap"
	"math"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// MillisecondBucket todo use timeout to calculate( if timeout smaller MillisecondBucket use it!)
	MillisecondBucket            int = 5000
	bucketNum                        = 9
	PopWaitSecond                    = 10 * time.Second
	EmptyPopSleepTime                = 50 * time.Millisecond
	AggregationEmptyPopSleepTime     = 500 * time.Millisecond
	ReportGoroutineNum               = 3
	KST                              = 3
)

type IntegratorUsecase struct {
	queRepository      repo.QueRepository
	reporterRepository ReporterRepository
	recordRepository   repo.RecordRepository
}

func NewIntegratorUsecase(queRepository repo.QueRepository, reporterRepository ReporterRepository,
	recordRepository repo.RecordRepository) *IntegratorUsecase {
	return &IntegratorUsecase{
		queRepository:      queRepository,
		reporterRepository: reporterRepository,
		recordRepository:   recordRepository,
	}
}

func (s *IntegratorUsecase) IntegrateReport(ctx context.Context, b *Integrate) error {
	logc.Info(ctx, "start integrate report", zap.Uint64("plan_id", b.PlanId))
	go s.rateReportResults(b)
	go s.aggregationResult(b)
	return nil
}

func (s *IntegratorUsecase) rateReportResults(in *Integrate) {
	ctx := context.Background()
	var wg sync.WaitGroup
	for _, task := range in.Tasks {
		task := task
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.calcTask(ctx, &task, in.PlanId, int(in.StressTime))
		}()
	}
	wg.Wait()
	logc.Info(ctx, "all task receive results summary success", zap.Uint64("plan_id", in.PlanId))
}

func (s *IntegratorUsecase) calcTask(ctx context.Context, task *Task, planId uint64, stressTime int) {
	start := time.Now()
	poolFunc, err := ants.NewPoolWithFunc(ReportGoroutineNum, func(data interface{}) {
		bar, ok := data.(*repo.Aggregate)
		if !ok {
			return
		}
		timeBuckets := make([]int64, MillisecondBucket+1)
		for duration, cnt := range bar.DurationMap {
			timeBuckets[duration] += cnt
		}
		latencyDistribution := calculateLatencyDistribution(bar.TotalNum, timeBuckets)
		latencyMap := make(map[string]int32)
		for _, lat := range latencyDistribution {
			latencyMap[fmt.Sprintf("%.2f", lat.Percentage)] = int32(lat.Latency)
		}
		err := s.reporterRepository.Report(ctx, []*Report{
			{
				PlanId:                     planId,
				TaskId:                     task.TaskId,
				Url:                        task.Url,
				Timestamp:                  bar.Timestamp,
				TotalNum:                   bar.TotalNum,
				TotalResponseContentLength: bar.TotalResponseContentLength,
				DurationMap:                bar.DurationMap,
				StatusMap:                  bar.StatusMap,
				ErrorMap:                   bar.ErrorMap,
				BodyCheckResultMap:         bar.BodyCheckResultMap,
				LatencyMap:                 latencyMap,
			}})
		if err != nil {
			logc.Error(ctx, "report error", zap.Error(err))
		}
	})
	tars := make(map[int64]*repo.Aggregate, KST)
	for {
		result, err := s.queRepository.RatePop(ctx, task.TaskId)
		if time.Since(start) > time.Second*time.Duration(stressTime)+PopWaitSecond {
			logc.Info(ctx, "pop task break...", zap.Uint64("TaskId", task.TaskId))
			break
		}
		if result == nil {
			time.Sleep(EmptyPopSleepTime)
			continue
		}
		if err != nil {
			logc.Error(ctx, "pop task error", zap.Error(err))
			continue
		}
		timestamp := result.Timestamp
		var car *repo.Aggregate
		var ok bool
		if car, ok = tars[timestamp]; !ok {
			car = repo.NewAggeRate()
			car.Timestamp = timestamp
			tars[timestamp] = car
		}
		car.TotalNum += result.TotalNum
		car.TotalResponseContentLength += result.TotalResponseContentLength
		for duration, cnt := range result.DurationMap {
			car.DurationMap[duration] += cnt
		}
		for statusCode, cnt := range result.StatusMap {
			car.StatusMap[statusCode] += cnt
		}
		for errStr, cnt := range result.ErrorMap {
			car.ErrorMap[errStr] += cnt
		}
		for bodyCheckResult, cnt := range result.BodyCheckResultMap {
			car.BodyCheckResultMap[bodyCheckResult] += cnt
		}
		if len(tars) > KST {
			minTimestamp := int64(math.MaxInt64)
			var bar *repo.Aggregate
			for ts, ar := range tars {
				if ts < minTimestamp {
					minTimestamp = ts
					bar = ar
				}
			}
			err := poolFunc.Invoke(bar)
			if err != nil {
				logc.Error(ctx, "poolFunc.Invoke error", zap.Error(err))
			}
			delete(tars, minTimestamp)
		}
		if result.Stop {
			logc.Info(ctx, "pop task stop break...", zap.Uint64("TaskId", task.TaskId))
			break
		}
	}
	for _, bar := range tars {
		err := poolFunc.Invoke(bar)
		if err != nil {
			logc.Error(ctx, "poolFunc.Invoke error", zap.Error(err))
		}
	}
	err = s.queRepository.RateClear(ctx, task.TaskId)
	poolFunc.Release()
	if err != nil {
		logc.Error(ctx, "clear task queue error", zap.Error(err))
	}
}

func (s *IntegratorUsecase) updateTaskRecord(ctx context.Context, taskId uint64, r []Summary) error {
	current := time.Now().Unix()
	marshal, err := json.Marshal(r)
	if err != nil {
		return err
	}
	taskRecord := &repo.TaskRecord{
		UpdateTime: current,
		TaskStatus: int(enums.DONE),
		StatExt:    string(marshal),
	}
	return s.recordRepository.UpdateTaskById(ctx, taskId, taskRecord)
}

func (s *IntegratorUsecase) aggregationResult(in *Integrate) {
	ctx := context.Background()
	var wg sync.WaitGroup
	for _, task := range in.Tasks {
		task := task
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.aggregationStat(ctx, &task, in)
		}()
	}
	wg.Wait()
	logc.Info(ctx, "all task receive results aggregation success")
	err := s.recordRepository.UpdateTaskById(ctx, in.PlanId, &repo.TaskRecord{TaskStatus: int(enums.DONE), UpdateTime: time.Now().Unix()})
	if err != nil {
		logc.Error(ctx, "update plan record error", zap.Error(err))
	}
}

func (s *IntegratorUsecase) aggregationStat(ctx context.Context, task *Task, in *Integrate) {
	intervalLen := int(in.IntervalLen)
	stressTime := int(in.StressTime)
	planId := in.PlanId
	rs := make([]Summary, intervalLen)
	for i := range rs {
		rs[i].TaskId = task.TaskId
		rs[i].Fastest = int64(MillisecondBucket)
		rs[i].ErrorDist = make(map[string]int64)
		rs[i].StatusCodeDist = make(map[int]int64)
		rs[i].TimeBuckets = make([]int64, MillisecondBucket+1)
		rs[i].BodyCheckResultMap = make(map[string]int64)
	}
	start := time.Now()
	for {
		result, err := s.queRepository.AggregatePop(ctx, task.TaskId)
		if time.Since(start) > time.Second*time.Duration(stressTime)+PopWaitSecond {
			logc.Info(ctx, "aggregation pop task break...", zap.Uint64("TaskId", task.TaskId))
			break
		}
		if result == nil {
			time.Sleep(AggregationEmptyPopSleepTime)
			continue
		}
		if err != nil {
			logc.Error(ctx, "pop task error", zap.Error(err))
			continue
		}
		i := result.Interval
		for costMs, cnt := range result.DurationMap {
			if cnt < 0 {
				continue
			}
			atomic.AddInt64(&rs[i].NumRes, cnt)
			du := int64(costMs)
			rs[i].Slowest = max(rs[i].Slowest, du)
			rs[i].Fastest = min(rs[i].Fastest, du)
			atomic.AddInt64(&rs[i].TimeBuckets[du], cnt)
			atomic.AddInt64(&rs[i].AvgTotal, du*cnt)
		}
		for errorInfo, cnt := range result.ErrorMap {
			if len(errorInfo) > 0 {
				atomic.AddInt64(&rs[i].ErrorCount, cnt)
				rs[i].ErrorDist[errorInfo] += cnt
			}
		}
		for statusCode, cnt := range result.StatusMap {
			rs[i].StatusCodeDist[int(statusCode)] += cnt
		}
		for checkResult, cnt := range result.BodyCheckResultMap {
			rs[i].BodyCheckResultMap[checkResult] += cnt
		}
		if result.TotalResponseContentLength > 0 {
			atomic.AddInt64(&rs[i].SizeTotal, result.TotalResponseContentLength)
		}
		if result.Stop {
			logc.Info(ctx, "aggregation pop task stop...", zap.Uint64("TaskId", task.TaskId))
			break
		}
	}
	err := s.queRepository.AggregateClear(ctx, task.TaskId)
	logc.Info(ctx, "aggregation task collect done, start to calculate", zap.Uint64("PlanId", planId), zap.Uint64("TaskId", task.TaskId))
	for i := range rs {
		if int32(enums.Step) == in.StressMode {
			latencyCalculate(&rs[i], int(in.StepIntervalTime))
		} else {
			latencyCalculate(&rs[i], stressTime)
		}
	}
	// write record
	err = s.updateTaskRecord(ctx, task.TaskId, rs)
	if err != nil {
		logc.Error(ctx, "update task record error", zap.Error(err))
	}
}

// latencyCalculate calculate latency distribution and buckets histogram
func latencyCalculate(r *Summary, time int) {
	r.TotalCostTime = float64(time)
	r.Rps = formatDecimal(float64(r.NumRes) / r.TotalCostTime) // actual using all request response time?
	r.Average = formatDecimal(float64(r.AvgTotal / r.NumRes))  // avg cost time

	r.LatencyDistribution = calculateLatencyDistribution(r.NumRes, r.TimeBuckets)
	buckets := make([]int, bucketNum+1)
	counts := make([]int64, bucketNum+1)
	bs := float64(r.Slowest-r.Fastest) / float64(bucketNum)
	for i := 0; i < bucketNum; i++ {
		buckets[i] = int(float64(r.Fastest) + bs*float64(i))
	}
	buckets[bucketNum] = int(r.Slowest)
	var bi int
	for ms := 0; ms <= MillisecondBucket; {
		if ms <= buckets[bi] {
			counts[bi] += r.TimeBuckets[ms]
			ms++
		} else if bi < len(buckets)-1 {
			bi++
		} else {
			break
		}
	}
	r.Histogram = make([]Bucket, len(buckets))
	totalDecimal := float64(0)
	for i := 0; i < len(buckets); i++ {
		decimal := formatDecimal(float64(counts[i]) / float64(r.NumRes))
		totalDecimal += decimal
		r.Histogram[i] = Bucket{
			Mark:      buckets[i],
			Count:     counts[i],
			Frequency: decimal,
		}
	}
}

// calculateLatencyDistribution calculate latency distribution with time buckets and total number of responses
func calculateLatencyDistribution(totalNumRes int64, timeBuckets []int64) []LatencyDistribution {
	pcs := []float64{100, 250, 500, 750, 900, 950, 990, 999}
	data := make([]int, len(pcs))
	cur := int64(0)
	idx := 0
	for ms := 0; ms <= MillisecondBucket; ms++ {
		if timeBuckets[ms] <= 0 {
			continue
		}
		cur += timeBuckets[ms]
		for idx < len(pcs) && float64(cur)*float64(1000)/float64(totalNumRes) >= pcs[idx] {
			data[idx] = ms
			idx++
		}
	}
	latencyDistribution := make([]LatencyDistribution, len(pcs))
	for i := 0; i < len(pcs); i++ {
		if data[i] > 0 {
			latencyDistribution[i] = LatencyDistribution{Percentage: float64(pcs[i]) / 10,
				Latency:       data[i],
				PercentageStr: fmt.Sprintf("%.1f", formatDecimal(float64(pcs[i])/10)) + "%",
				LatencyMs:     strconv.Itoa(data[i]) + "ms",
			}
		}
	}
	return latencyDistribution
}

func formatDecimal(value float64) float64 {
	value, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", value), 64)
	return value
}
