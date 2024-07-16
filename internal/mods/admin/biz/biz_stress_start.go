package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jinzhu/copier"
	integratorv1 "github.com/peckfly/gopeck/api/integrator/v1"
	peckv1 "github.com/peckfly/gopeck/api/pecker/v1"
	"github.com/peckfly/gopeck/internal/conf"
	"github.com/peckfly/gopeck/internal/mods/common/repo"
	"github.com/peckfly/gopeck/internal/pkg/consts"
	"github.com/peckfly/gopeck/internal/pkg/enums"
	"github.com/peckfly/gopeck/pkg/interpreter"
	"github.com/peckfly/gopeck/pkg/log/logc"
	"github.com/peckfly/gopeck/pkg/netx"
	"github.com/peckfly/gopeck/pkg/numx"
	"github.com/peckfly/gopeck/pkg/registry"
	"github.com/peckfly/gopeck/pkg/registry/discovery"
	"github.com/spf13/cast"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	grpcinsecure "google.golang.org/grpc/credentials/insecure"
	"net/url"
	"strings"
	"time"
)

type StressUsecase struct {
	recordRepository repo.RecordRepository
	nodeRepository   repo.NodeRepository
	queRepository    repo.QueRepository
	discovery        registry.Discovery
	metricsConf      conf.MetricsConf
	stressConf       conf.WorkerStressConf
}

// NewStressUsecase new stress usecase
func NewStressUsecase(recordRepository repo.RecordRepository, nodeRepository repo.NodeRepository,
	queRepository repo.QueRepository, discovery registry.Discovery, conf *conf.ServerConf) *StressUsecase {
	return &StressUsecase{
		recordRepository: recordRepository,
		nodeRepository:   nodeRepository,
		queRepository:    queRepository,
		discovery:        discovery,
		metricsConf:      conf.Metrics,
		stressConf:       conf.StressConf,
	}
}

// StartStress starts the stress use case.
//
// ctx: context.Context, in: *Plan
// error
func (s *StressUsecase) StartStress(ctx context.Context, in *Plan) error {
	err := s.preCheck(ctx, in)
	if err != nil {
		return err
	}
	services, err := s.discovery.GetService(ctx, consts.Pecker)
	if err != nil {
		return err
	}
	nodeInfos, err := s.nodeRepository.GetAllNodeInfo(ctx)
	if err != nil {
		return err
	}
	nodeCostInfos := getNodeCostInfoByInstances(services, nodeInfos)
	planId, err := generatePlanId()
	if err != nil {
		return err
	}
	in.PlanId = planId
	in.StressTime *= int(time.Minute / time.Second)
	in.StepIntervalTime *= int(time.Minute / time.Second)
	var updateNodes [][2]*repo.Node
	stressModeType := enums.StressModeType(in.StressMode)
	isStepMode := stressModeType == enums.Step
	intervalLen := 1 // adaptive with default mode
	if isStepMode {
		div := numx.CeilDiv(in.StressTime, in.StepIntervalTime)
		intervalLen = max(intervalLen, div)
	} else {
		// adaptive with default mode
		in.StepIntervalTime = in.StressTime
	}
	in.IntervalLen = intervalLen
	err = s.assignAndCalculateTask(planId, in, isStepMode, intervalLen, nodeCostInfos, updateNodes)
	if err != nil {
		return err
	}
	// build connection
	connMap := make(map[uint64][]*BindConn)
	for _, task := range in.Tasks {
		for _, node := range task.nodes {
			var conn *grpc.ClientConn
			conn, err = grpc.DialContext(ctx, strings.ReplaceAll(node.nodeInfo.Addr, "grpc://", ""), grpc.WithTransportCredentials(grpcinsecure.NewCredentials()))
			if err != nil {
				logc.Error(ctx, "failed to dial node", zap.String("addr", node.nodeInfo.Addr), zap.Error(err))
				return err
			}
			connMap[task.TaskId] = append(connMap[task.TaskId], &BindConn{node.num, node.Nums, node.nodeInfo.Addr, conn})
		}
	}
	err = s.insertPlanTaskRecord(ctx, in)
	if err != nil {
		return err
	}
	// start to receive results
	err = s.receiveResults(ctx, in)
	if err != nil {
		return err
	}
	// todo: if one or more nodes failed to sendRequest, deal with it
	for taskId, conns := range connMap {
		t := getTaskByTaskId(taskId, in.Tasks)
		s.sendRequest(ctx, conns, t)
	}
	err = s.nodeRepository.UpdateNodeInfoList(ctx, updateNodes)
	if err != nil {
		logc.Error(ctx, "failed to update node info", zap.Error(err))
		return err
	}
	return nil
}

// assignAndCalculateTask assign tasks to nodes and calculate the costs
func (s *StressUsecase) assignAndCalculateTask(planId uint64, in *Plan, isStepMode bool, intervalLen int, nodeCostInfos []*NodeInstanceCost, updateNodes [][2]*repo.Node) error {
	for i, task := range in.Tasks {
		taskId, err := generateTaskId()
		if err != nil {
			return err
		}
		s.taskParamSetting(in, i, taskId, planId)
		planStressType := enums.StressType(in.Tasks[i].StressType)
		curNum := task.Num
		if isStepMode {
			curNum = max(task.Num, task.MaxNum)
		}
		totNums := curNum
		costs := make([]int32, intervalLen)
		assigned := false
		for _, node := range nodeCostInfos {
			nodeMaxConcurrencyNum, nodeMaxRpsNum := node.GoroutineQuota, node.RpsQuota
			var leftNum int
			if node.RpsCost > 0 {
				if planStressType != enums.Rps {
					continue
				}
				leftNum = nodeMaxRpsNum - node.RpsCost
			} else if node.GoroutineCost > 0 {
				if planStressType != enums.Concurrency {
					continue
				}
				leftNum = nodeMaxConcurrencyNum - node.GoroutineCost
			} else {
				if planStressType == enums.Rps {
					leftNum = nodeMaxRpsNum
				} else {
					leftNum = nodeMaxConcurrencyNum
				}
			}
			addCostNum := min(curNum, leftNum)
			curNum -= addCostNum
			nums := make([]int32, intervalLen)
			if isStepMode {
				startNum := task.Num
				for j := 0; j < intervalLen; j++ {
					nums[j] = int32(startNum * addCostNum / totNums)
					if curNum <= 0 {
						nums[j] = int32(startNum) - costs[j]
					} else {
						costs[j] += nums[j]
					}
					startNum = min(startNum+task.StepNum, task.MaxNum)
				}
			}
			// updateStatus
			newNodeInfo := &repo.Node{
				Addr:             node.instance.Addr,
				RpsCost:          node.RpsCost,
				GoroutineCost:    node.GoroutineCost,
				RunningTaskCount: node.RunningTaskCount,
				RpsQuota:         node.RpsQuota,
				GoroutineQuota:   node.GoroutineQuota,
			}
			if planStressType == enums.Rps {
				newNodeInfo.RpsCost += addCostNum
			} else if planStressType == enums.Concurrency {
				newNodeInfo.GoroutineCost += addCostNum
			}
			newNodeInfo.RunningTaskCount += 1
			if node.UnExist {
				updateNodes = append(updateNodes, [2]*repo.Node{nil, newNodeInfo})
			} else {
				updateNodes = append(updateNodes, [2]*repo.Node{
					{
						Addr:             node.instance.Addr,
						RpsCost:          node.RpsCost,
						GoroutineCost:    node.GoroutineCost,
						RunningTaskCount: node.RunningTaskCount,
						RpsQuota:         node.RpsQuota,
						GoroutineQuota:   node.GoroutineQuota,
					},
					newNodeInfo,
				})
			}
			// send request to this address
			in.Tasks[i].nodes = append(in.Tasks[i].nodes, &BindNode{addCostNum, nums, newNodeInfo})
			// check left curNum
			if curNum <= 0 {
				assigned = true
				break
			}
		}
		if !assigned {
			return fmt.Errorf("nodes are busy, or nodes have no enough resources to run this task")
		}
	}
	return nil
}

// taskParamSetting sets the parameters of a task in a StressUsecase.
//
// Parameters:
// - in: the Plan struct pointer containing the task
// - i: the index of the task in the Tasks slice
// - taskId: the unique identifier of the task
// - planId: the unique identifier of the plan
func (s *StressUsecase) taskParamSetting(in *Plan, i int, taskId uint64, planId uint64) {
	in.Tasks[i].TaskId = taskId
	in.Tasks[i].PlanId = planId
	in.Tasks[i].StressTime = in.StressTime
	in.Tasks[i].StressType = in.StressType
	in.Tasks[i].StressMode = in.StressMode
	in.Tasks[i].StepIntervalTime = in.StepIntervalTime
	in.Tasks[i].Headers = parseEntryList(in.Tasks[i].HeaderEntry)
	in.Tasks[i].Query = parseEntryList(in.Tasks[i].QueryEntry)
	if in.Tasks[i].MaxConnections <= 0 {
		in.Tasks[i].MaxConnections = s.stressConf.DefaultMaxConnections
	}
	if in.Tasks[i].Timeout <= 0 {
		in.Tasks[i].Timeout = maxTimeoutSecond
	}
	in.Tasks[i].Timeout = min(in.Tasks[i].Timeout, maxTimeoutSecond)
	in.Tasks[i].DisableCompression = parseOtherOptions(in.Tasks[i].Options, disableCompression)
	in.Tasks[i].DisableKeepAlive = parseOtherOptions(in.Tasks[i].Options, disableKeepAlive)
	in.Tasks[i].DisableRedirects = parseOtherOptions(in.Tasks[i].Options, disableRedirects)
	in.Tasks[i].H2 = parseOtherOptions(in.Tasks[i].Options, h2)
}

// getNodeCostInfoByInstances generates NodeInstanceCost based on services and Node info.
//
// Parameter(s):
// - services: slice of registry.ServiceInstance pointers
// - infos: slice of repo.Node pointers
// Return type(s):
// - slice of NodeInstanceCost pointers
func getNodeCostInfoByInstances(services []*registry.ServiceInstance, infos []*repo.Node) []*NodeInstanceCost {
	costNodeInfos := make([]*NodeInstanceCost, 0)
	for _, service := range services {
		nodeInfo := &repo.Node{
			Addr:             service.Addr,
			RpsQuota:         cast.ToInt(service.Metadata[consts.MaxRpsNum]),
			GoroutineQuota:   cast.ToInt(service.Metadata[consts.MaxConcurrencyNum]),
			RunningTaskCount: 0,
			RpsCost:          0,
			GoroutineCost:    0,
		}
		exist := false
		for _, info := range infos {
			if service.Addr == info.Addr {
				nodeInfo = info
				exist = true
				break
			}
		}
		costNodeInfos = append(costNodeInfos, &NodeInstanceCost{
			instance:         service,
			RpsCost:          nodeInfo.RpsCost,
			GoroutineCost:    nodeInfo.GoroutineCost,
			RunningTaskCount: nodeInfo.RunningTaskCount,
			RpsQuota:         nodeInfo.RpsQuota,
			GoroutineQuota:   nodeInfo.GoroutineQuota,
			UnExist:          !exist,
		})
	}
	return costNodeInfos
}

// preCheck checks the validity of the stress plan before execution.
//
// It takes a context and a Plan pointer as input parameters.
// It returns an error if the plan is invalid.
func (s *StressUsecase) preCheck(ctx context.Context, in *Plan) error {
	if !enums.SupportStressType(in.StressType) {
		return fmt.Errorf("not support stress type: %d", in.StressType)
	}
	if len(in.Tasks) > maxTaskCount {
		return fmt.Errorf("task count should be less than %d", maxTaskCount)
	}
	for i, task := range in.Tasks {
		if !strings.HasPrefix(task.Url, consts.HttpScheme) && !strings.HasPrefix(task.Url, consts.HttpsScheme) {
			return fmt.Errorf("URL: %s, should start with http:// or https://", task.Url)
		}
		_, err := url.Parse(task.Url)
		if err != nil {
			logc.Info(ctx, "parse url error", zap.String("url", task.Url), zap.Error(err))
			return fmt.Errorf("URL: %s, parse url error %+v", task.Url, err)
		}
		if err = netx.Ping(task.Url); err != nil {
			logc.Error(ctx, "ping url error", zap.String("url", task.Url), zap.Error(err))
			return err
		}
		if len(task.Body) > 0 && !json.Valid([]byte(task.Body)) {
			return fmt.Errorf("URL: %s, body should be json", task.Url)
		}
		if len(task.DynamicParamScript) > 0 {
			in.Tasks[i].DynamicParams, err = parseDynamicParams(task)
			if err != nil {
				return err
			}
		}
		if len(task.ResponseCheckScript) > 0 {
			err = checkResponseCheckScript(task.ResponseCheckScript)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// DeleteNodeStateInfo deletes node state information.
//
// ctx: The context in which the function is being called.
// addr: The address of the node state information to be deleted.
// error: An error, if any, that occurred during the deletion process.
func checkResponseCheckScript(responseCheckScript string) error {
	evalInterpreter, err := interpreter.NewEvalInterpreter(responseCheckScript, interpreter.WithFuncName("Check"))
	if err != nil {
		return err
	}
	err = evalInterpreter.ExecuteScript(func(executor any) {
		executor.(func(string) string)("")
	})
	if err != nil {
		return err
	}
	return nil
}

// sendRequest sends a peck request to multiple connections using the provided task.
//
// Parameters:
//   - ctx: the context for the request
//   - conns: slice of BindConn connections to send the request to
//   - task: the task to be executed
func (s *StressUsecase) sendRequest(ctx context.Context, conns []*BindConn, task Task) {
	var request peckv1.PeckRequest
	err := copier.Copy(&request, task)
	if err != nil {
		logc.Error(ctx, "copy task error", zap.Error(err))
	}
	request.Query = netx.ParseQueryMap(task.Query)
	err = copier.Copy(&request.DynamicParams, task.DynamicParams)
	if err != nil {
		logc.Error(ctx, "copy dynamic params error", zap.Error(err))
	}
	for _, conn := range conns {
		peckServiceClient := peckv1.NewPeckServiceClient(conn.grpcConn)
		request.Num = int32(conn.num)
		request.Nums = conn.Nums
		request.Addr = conn.Addr
		for retryTimes := 0; retryTimes < 3; retryTimes++ {
			peckReply, err := peckServiceClient.Peck(ctx, &request)
			if err != nil {
				logc.Error(ctx, "peck request failed", zap.Error(err))
				continue
			}
			logc.Info(ctx, "peck request success", zap.Any("peck", peckReply))
			break
		}
		err = conn.grpcConn.Close()
		if err != nil {
			logc.Error(ctx, "close conn error", zap.Error(err))
		}
	}
}

// receiveResults is a function that processes the results received from the integration service.
//
// ctx: the context for the function
// in: the plan containing tasks to be integrated
// error: an error if any occurred during the integration process
func (s *StressUsecase) receiveResults(ctx context.Context, in *Plan) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()
	grpcClientConn, err := discovery.Dial(ctx, fmt.Sprintf("%s:///%s", discovery.Scheme, consts.Integrator), discovery.WithDiscovery(s.discovery))
	defer func() {
		err = grpcClientConn.Close()
		if err != nil {
			logc.Error(ctx, "close grpcClientConn error", zap.Error(err))
		}
	}()
	integrateServiceClient := integratorv1.NewIntegrateServiceClient(grpcClientConn)
	var tasks []*integratorv1.Task
	for _, task := range in.Tasks {
		tasks = append(tasks, &integratorv1.Task{
			Url:                  task.Url,
			TaskId:               task.TaskId,
			RequestContentLength: int32(len(task.Body)),
		})
	}
	integrateReply, err := integrateServiceClient.Integrate(ctx, &integratorv1.IntegrateRequest{
		PlanId:           in.PlanId,
		IntervalLen:      int32(in.IntervalLen),
		StressTime:       int32(in.StressTime),
		StartTime:        in.StartTime,
		StressType:       int32(in.StressType),
		StressMode:       int32(in.StressMode),
		StepIntervalTime: int32(in.StepIntervalTime),
		UserId:           in.UserId,
		Tasks:            tasks,
	})
	logc.Info(ctx, "integrate request success", zap.Any("integrateReply", integrateReply))
	return err
}

// getOverviewMetricsUrl generates the metrics URL for stress overview based on the plan ID, start time, and stress time.
//
// ctx is the context object, planId is the unique identifier for the plan, stressStartTime is the start time of stress in Unix timestamp,
// stressTime is the duration of stress in seconds.
// string - returns the generated metrics URL.
func (s *StressUsecase) getOverviewMetricsUrl(ctx context.Context, planId uint64, stressStartTime int64, stressTime int) string {
	metricsUrl, err := netx.AddUrlParam(s.metricsConf.GrafanaAddr, map[string]string{
		"var-planId": cast.ToString(planId),
		"from":       cast.ToString((stressStartTime - int64(60)) * 1000),
		"to":         cast.ToString((stressStartTime + int64(stressTime) + int64(60)) * 1000),
	})
	if err != nil {
		logc.Error(ctx, "add metrics url param error", zap.Error(err))
		return s.metricsConf.GrafanaAddr
	}
	return metricsUrl
}

// getTaskMetricsUrl generates the metrics URL for a specific task based on the provided parameters.
//
// ctx: context.Context - the context for the operation.
// planId: uint64 - the ID of the plan.
// taskId: uint64 - the ID of the task.
// stressStartTime: int64 - the start time of stress in Unix timestamp format.
// stressTime: int - the duration of stress in seconds.
// string - returns the generated metrics URL.
func (s *StressUsecase) getTaskMetricsUrl(ctx context.Context, planId uint64, taskId uint64, stressStartTime int64, stressTime int) string {
	metricsUrl, err := netx.AddUrlParam(s.metricsConf.GrafanaAddr, map[string]string{
		"var-planId": cast.ToString(planId),
		"var-taskId": cast.ToString(taskId),
		"from":       cast.ToString((stressStartTime - int64(60)) * 1000),
		"to":         cast.ToString((stressStartTime + int64(stressTime) + int64(60)) * 1000),
	})
	if err != nil {
		logc.Error(ctx, "add metrics url param error", zap.Error(err))
		return s.metricsConf.GrafanaAddr
	}
	return metricsUrl
}

// parseDynamicParams parses dynamic parameters for a given task.
//
// task Task
// []DynamicParam, error
func parseDynamicParams(task Task) ([]DynamicParam, error) {
	evalInterpreter, err := interpreter.NewEvalInterpreter(task.DynamicParamScript, interpreter.WithFuncName("GetParams"))
	if err != nil {
		return nil, err
	}
	var params string
	err = evalInterpreter.ExecuteScript(func(executor any) {
		params = executor.(func() string)()
	})
	if err != nil {
		return nil, err
	}
	if len(params) <= 0 {
		return nil, nil
	}
	if len(params) > maxDynamicParamLength {
		return nil, fmt.Errorf("dynamic param string length should be less than %d", maxDynamicParamLength)
	}
	var dynamicParams []DynamicParam
	err = json.Unmarshal([]byte(params), &dynamicParams)
	if err != nil {
		return nil, err
	}
	return dynamicParams, nil
}

// getTaskByTaskId retrieves a task with a specific ID from a slice of tasks.
//
// Parameters:
//
//	id - the ID of the task to retrieve
//	tasks - a slice of tasks to search in
//
// Return:
//
//	t - the task with the specified ID, if found
func getTaskByTaskId(id uint64, tasks []Task) (t Task) {
	for _, task := range tasks {
		if task.TaskId == id {
			t = task
			break
		}
	}
	return
}

// parseOtherOptions checks if the given compression option exists in the options slice.
//
// Parameters:
// - options []string: a slice of options to search through.
// - compression string: the compression option to look for.
// Returns a boolean indicating whether the compression option exists in the options slice.
func parseOtherOptions(options []string, compression string) bool {
	for _, option := range options {
		if option == compression {
			return true
		}
	}
	return false
}

// parseEntryList generates a map of string key-value pairs from a list of Entry structs.
//
// Parameter: entryList []Entry
// Return type: map[string]string
func parseEntryList(entryList []Entry) map[string]string {
	if len(entryList) == 0 {
		return nil
	}
	entryMap := make(map[string]string)
	for _, entry := range entryList {
		entryMap[entry.EntryKey] = entry.EntryValue
	}
	return entryMap
}
