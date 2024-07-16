package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jinzhu/copier"
	"github.com/peckfly/gopeck/internal/mods/common/repo"
	"github.com/peckfly/gopeck/internal/pkg/enums"
	"github.com/peckfly/gopeck/pkg/log/logc"
	"github.com/peckfly/gopeck/pkg/netx"
	"github.com/spf13/cast"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"time"
)

func (s *StressUsecase) insertPlanTaskRecord(ctx context.Context, in *Plan) error {
	current := time.Now().Unix()
	stressStartTime := in.StartTime
	// current support start stress immediately
	if in.StartTime <= 0 {
		stressStartTime = current
	}
	pr := &repo.PlanRecord{
		PlanId:           in.PlanId,
		UserId:           in.UserId,
		Status:           int(enums.DOING),
		PlanName:         in.PlanName,
		StressType:       in.StressType,
		StressMode:       in.StressMode,
		StressTime:       in.StressTime,
		StartTime:        stressStartTime,
		StepIntervalTime: in.StepIntervalTime,
		IntervalLen:      in.IntervalLen,
		CreateTime:       current,
		UpdateTime:       current,
	}
	err := s.recordRepository.CreatePlan(ctx, pr)
	if err != nil {
		return err
	}
	var taskRecords []repo.TaskRecord
	for _, task := range in.Tasks {
		var taskRecord repo.TaskRecord
		err = copier.Copy(&taskRecord, task)
		if err != nil {
			logc.Error(ctx, "copy task error", zap.Error(err))
			continue
		}
		taskRecord.Query = netx.ParseQueryMap(task.Query)
		headers, err := json.Marshal(task.Headers)
		if err != nil {
			logc.Error(ctx, "marshal headers error", zap.Error(err))
			continue
		}
		taskRecord.StressMode = in.StressMode
		taskRecord.Header = string(headers)
		taskRecord.CreateTime = current
		taskRecord.DisableCompression = cast.ToInt8(task.DisableCompression)
		taskRecord.DisableRedirects = cast.ToInt8(task.DisableRedirects)
		taskRecord.DisableKeepAlive = cast.ToInt8(task.DisableKeepAlive)
		taskRecord.TaskStatus = int(enums.DOING)
		if len(task.HeaderEntry) > 0 {
			headerStr, err := json.Marshal(task.HeaderEntry)
			if err == nil {
				taskRecord.Header = string(headerStr)
			}
		}
		if len(task.QueryEntry) > 0 {
			queryStr, err := json.Marshal(task.QueryEntry)
			if err == nil {
				taskRecord.Query = string(queryStr)
			}
		}
		taskRecord.H2 = cast.ToInt8(task.H2)
		var nodeAddrs []string
		for _, node := range task.nodes {
			nodeAddrs = append(nodeAddrs, node.nodeInfo.Addr)
		}
		taskRecord.Nodes = strings.Join(nodeAddrs, ",")
		taskRecords = append(taskRecords, taskRecord)
	}
	err = s.recordRepository.BatchCreateTasks(ctx, taskRecords)
	if err != nil {
		return err
	}
	return nil
}

func (s *StressUsecase) QueryUserRecords(ctx context.Context, userId string, query PlanRecordQuery) (*PlanQueryResult, error) {
	query.Pagination = true
	records, pp, err := s.recordRepository.QueryUserRecordsByUserId(ctx, userId, query.PlanId, query.PlanName, query.StartTime, query.EndTime,
		query.PaginationParam)
	if err != nil {
		return nil, err
	}
	planResults := make([]*PlanRecordResultItem, len(records))
	current := time.Now().Unix()
	for i := range records {
		planResults[i] = new(PlanRecordResultItem)
		copier.Copy(planResults[i], records[i])
		planResults[i].PlanId = strconv.FormatUint(records[i].PlanId, 10)
		stressStartTime := planResults[i].StartTime
		stressTime := planResults[i].StressTime
		if planResults[i].Status == int(enums.DONE) || planResults[i].Status == int(enums.STOP) {
			planResults[i].StressProgress = 100
		} else {
			progress := fmt.Sprintf("%.1f", (float64(current)-float64(stressStartTime))/float64(stressTime+PopWaitSecond)*100)
			finalProgress, err := strconv.ParseFloat(progress, 64)
			if err == nil {
				planResults[i].StressProgress = min(100, finalProgress)
			}
		}
		planResults[i].OverviewMetricsUrl = s.getOverviewMetricsUrl(ctx, records[i].PlanId, stressStartTime, stressTime)
		planResults[i].StressTime /= int(time.Minute / time.Second)
	}
	return &PlanQueryResult{
		Data:       planResults,
		PageResult: pp,
	}, nil
}

func (s *StressUsecase) QueryPlanTaskRecords(ctx context.Context, userId string, query TaskRecordQuery) (*TaskResult, error) {
	planId, err := strconv.ParseUint(query.PlanId, 10, 64)
	if err != nil {
		return nil, err
	}
	logc.Info(ctx, "query plan task records", zap.String("userId", userId), zap.Uint64("planId", planId))
	records, err := s.recordRepository.QueryTaskRecordsByPlanId(ctx, planId)
	if err != nil {
		return nil, err
	}
	items := make([]*TaskResultItem, len(records))
	planRecord, err := s.recordRepository.FindPlanRecordByPlanId(ctx, planId)
	if err != nil {
		return nil, err
	}
	for i := range records {
		items[i] = new(TaskResultItem)
		copier.Copy(items[i], records[i])
		items[i].PlanId = strconv.FormatUint(records[i].PlanId, 10)
		items[i].TaskId = strconv.FormatUint(records[i].TaskId, 10)
		items[i].HeaderEntry = formatEntryList(records[i].Header)
		items[i].QueryEntry = formatEntryList(records[i].Query)
		if len(records[i].Body) > 0 {
			items[i].Body = &JsonBody{}
			err := json.Unmarshal([]byte(records[i].Body), &items[i].Body.Json)
			if err != nil {
				logc.Error(ctx, "unmarshal body error", zap.Error(err))
			}
		}
		items[i].MetricsUrl = s.getTaskMetricsUrl(ctx, records[i].PlanId, records[i].TaskId, planRecord.StartTime, planRecord.StressTime)
		if len(records[i].StatExt) > 0 {
			err := json.Unmarshal([]byte(records[i].StatExt), &items[i].Reports)
			if err != nil {
				logc.Error(ctx, "unmarshal statExt error", zap.Error(err))
			}
			startNum := items[i].Num
			for j := range items[i].Reports {
				items[i].Reports[j].Num = startNum
				startNum += items[i].StepNum
				items[i].Reports[j].Lat90 = findLatencyPercent(items[i].Reports[j].LatencyDistribution, "90.0%")
				items[i].Reports[j].Lat95 = findLatencyPercent(items[i].Reports[j].LatencyDistribution, "95.0%")
				items[i].Reports[j].Lat99 = findLatencyPercent(items[i].Reports[j].LatencyDistribution, "99.0%")
				items[i].Reports[j].Lat999 = findLatencyPercent(items[i].Reports[j].LatencyDistribution, "99.9%")
				items[i].Reports[j].ErrorRate = fmt.Sprintf("%.2f", float64(items[i].Reports[j].ErrorCount)/float64(items[i].Reports[j].NumRes)) + "%"
			}
		}
	}
	return &TaskResult{Data: items}, nil
}

// QueryPlanAndTaskRecords query plan and task record to copy stress plan task
func (s *StressUsecase) QueryPlanAndTaskRecords(ctx context.Context, userId string, query TaskRecordQuery) (*Plan, error) {
	planId, err := strconv.ParseUint(query.PlanId, 10, 64)
	if err != nil {
		return nil, err
	}
	records, err := s.recordRepository.QueryTaskRecordsByPlanId(ctx, planId)
	if err != nil {
		return nil, err
	}
	logc.Info(ctx, "query plan and task records", zap.String("userId", userId), zap.Uint64("planId", planId))
	planRecord, err := s.recordRepository.FindPlanRecordByPlanId(ctx, planId)
	if err != nil {
		return nil, err
	}
	planResult := new(Plan)
	copier.Copy(planResult, planRecord)
	planResult.StartTime = 0 //reset startTime
	planResult.StressTime /= int(time.Minute / time.Second)
	tasks := make([]Task, len(records))
	for i := range records {
		copier.Copy(&tasks[i], records[i])
		tasks[i].QueryEntry = getFromEntryList(records[i].Query)
		tasks[i].HeaderEntry = getFromEntryList(records[i].Header)
		tasks[i].Options = getTaskOption(records[i])
	}
	planResult.Tasks = tasks
	return planResult, nil
}

func getFromEntryList(entryStr string) []Entry {
	var entryList []Entry
	err := json.Unmarshal([]byte(entryStr), &entryList)
	if err != nil {
		return nil
	}
	return entryList
}

func getTaskOption(record *repo.TaskRecord) []string {
	var options []string
	if record.DisableKeepAlive == 1 {
		options = append(options, disableKeepAlive)
	}
	if record.DisableCompression == 1 {
		options = append(options, disableCompression)
	}
	if record.DisableRedirects == 1 {
		options = append(options, disableRedirects)
	}
	if record.H2 == 1 {
		options = append(options, h2)
	}
	return options
}

func findLatencyPercent(distributions []LatencyDistribution, percentStr string) int {
	if len(distributions) == 0 {
		return 0
	}
	for j := range distributions {
		if distributions[j].PercentageStr == percentStr {
			return distributions[j].Latency
		}
	}
	return 0
}

func formatEntryList(s string) []Entry {
	var list []Entry
	_ = json.Unmarshal([]byte(s), &list)
	return list
}
