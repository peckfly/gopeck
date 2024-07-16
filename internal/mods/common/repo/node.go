package repo

import (
	"context"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/process"
)

type (
	Node struct {
		// cost quota in this node
		Addr             string `json:"addr"`
		RpsCost          int    `json:"rps_cost"`
		GoroutineCost    int    `json:"goroutine_cost"`
		RpsQuota         int    `json:"rps_quota"`
		GoroutineQuota   int    `json:"goroutine_quota"`
		RunningTaskCount int    `json:"running_task_count"`
	}

	NodeState struct {
		Addr         string
		Timestamp    int64
		LoadInfo     *load.AvgStat
		RunTimeInfo  RunTimeInfo
		ProgressInfo ProgressInfo
		MemInfoList  []MemInfo
	}

	RunTimeInfo struct {
		GoRoutineNum int
	}

	ProgressInfo struct {
		CpuCount   int
		MemPercent float32
		CPUPercent float64
		MemoryStat *process.MemoryInfoStat
	}
	MemInfo struct {
		Total       uint64  `json:"total"`
		Used        uint64  `json:"used"`
		Free        uint64  `json:"free"`
		UsedPercent float64 `json:"usedPercent"`
	}
	DiskInfo struct {
		Total       uint64  `json:"total"`
		Free        uint64  `json:"free"`
		Used        uint64  `json:"used"`
		UsedPercent float64 `json:"usedPercent"`
	}
	Network struct {
		Name        string `json:"name"`
		BytesSent   uint64 `json:"bytesSent"`
		BytesRecv   uint64 `json:"bytesRecv"`
		PacketsSent uint64 `json:"packetsSent"`
		PacketsRecv uint64 `json:"packetsRecv"`
	}

	NodeRepository interface {
		// UpdateNodeInfoList update node info list
		UpdateNodeInfoList(ctx context.Context, infos [][2]*Node) error
		// GetAllNodeInfo get all node info
		GetAllNodeInfo(ctx context.Context) ([]*Node, error)
		// DeleteNodeInfo delete node info
		DeleteNodeInfo(ctx context.Context, addr string) error
		// UpdateNodeCostNum update node cost num
		UpdateNodeCostNum(ctx context.Context, addr string, stressType, num int) error
		// UpdateNodeQuota update node quota
		UpdateNodeQuota(ctx context.Context, addr string, rpsQuota, goroutineQuota int) error
		// ReportNodeInfo report node info
		ReportNodeInfo(ctx context.Context, state *NodeState) error
		// DeleteNodeStateInfo delete node state info
		DeleteNodeStateInfo(ctx context.Context, addr string) error
		// BatchGetNodeState batch get node state
		BatchGetNodeState(ctx context.Context, addrs []string) ([]*NodeState, error)
	}
)
