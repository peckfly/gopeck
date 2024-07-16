package server

import (
	"github.com/peckfly/gopeck/internal/mods/common/repo"
	"github.com/peckfly/gopeck/pkg/log"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
	"go.uber.org/zap"
	"os"
	"runtime"
	"time"
)

// GetNodeState returns the node status including progress, disk information, memory information, and network information.
//
// Returns a pointer to repo.NodeState.
func GetNodeState() (nodeStatus *repo.NodeState) {
	nodeStatus = &repo.NodeState{}
	nodeStatus.ProgressInfo = GetProgressInfo()
	nodeStatus.MemInfoList = GetMemInfo()
	nodeStatus.LoadInfo = GetCPULoad()
	nodeStatus.RunTimeInfo = GetRuntimeInfo()
	nodeStatus.Timestamp = time.Now().Unix()
	return
}

// GetRuntimeInfo retrieves the runtime information.
//
// Returns a RunTimeInfo struct.
func GetRuntimeInfo() (runTimeInfo repo.RunTimeInfo) {
	runTimeInfo = repo.RunTimeInfo{}
	runTimeInfo.GoRoutineNum = runtime.NumGoroutine()
	return
}

// GetProgressInfo retrieves progress information.
//
// No parameters.
// Returns a repo.ProgressInfo struct.
func GetProgressInfo() (progressInfo repo.ProgressInfo) {
	progressInfo = repo.ProgressInfo{}
	var err error
	p, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		log.Error("Error creating process instance:", zap.Error(err))
		return
	}
	progressInfo.CpuCount, err = cpu.Counts(true)
	if err != nil {
		log.Error("Error getting process CPU count:", zap.Error(err))
		return
	}
	progressInfo.CPUPercent, err = p.CPUPercent()
	if err != nil {
		log.Error("Error getting process CPU percent:", zap.Error(err))
		return
	}
	progressInfo.MemPercent, err = p.MemoryPercent()
	if err != nil {
		log.Error("Error getting process memory percent:", zap.Error(err))
		return
	}
	progressInfo.MemoryStat, err = p.MemoryInfo()

	if err != nil {
		log.Error("Error getting process memory info:", zap.Error(err))
		return
	}
	return
}

// GetCPULoad retrieves the CPU load information.
//
// No parameters.
// Returns *load.AvgStat.
func GetCPULoad() (info *load.AvgStat) {
	info, err := load.Avg()
	if err != nil {
		return
	}
	return info
}

// GetMemInfo retrieves memory information and returns a list of MemInfo.
//
// No parameters.
// Returns a list of MemInfo.
func GetMemInfo() (memInfoList []repo.MemInfo) {
	memVir := repo.MemInfo{}
	memInfoVir, err := mem.VirtualMemory()
	if err != nil {
		return
	}
	memVir.Total = memInfoVir.Total / 1024 / 1024 / 1024
	memVir.Free = memInfoVir.Free / 1024 / 1024
	memVir.Used = memInfoVir.Used / 1024 / 1024
	memVir.UsedPercent = memInfoVir.UsedPercent
	memInfoList = append(memInfoList, memVir)

	memInfoSwap, err := mem.SwapMemory()
	if err != nil {
		return
	}
	memVir.Total = memInfoSwap.Total / 1024 / 1024 / 1024
	memVir.Free = memInfoSwap.Free / 1024 / 1024
	memVir.Used = memInfoSwap.Used / 1024 / 1024
	memVir.UsedPercent = memInfoSwap.UsedPercent
	memInfoList = append(memInfoList, memVir)
	return memInfoList
}

// GetHostName returns the hostname of the host machine.
//
// No parameters.
// Returns a string.
func GetHostName() string {
	hostInfo, err := host.Info()
	if err != nil {
		return ""
	}
	return hostInfo.Hostname
}

// GetDiskInfo retrieves disk information and returns a list of repo.DiskInfo.
//
// No parameters.
// Returns a slice of repo.DiskInfo.
func GetDiskInfo() (diskInfoList []repo.DiskInfo) {
	disks, err := disk.Partitions(true)
	if err != nil {
		return
	}
	for _, v := range disks {
		diskInfo := repo.DiskInfo{}
		info, err := disk.Usage(v.Device)
		if err != nil {
			continue
		}
		diskInfo.Total = info.Total
		diskInfo.Free = info.Free
		diskInfo.Used = info.Used
		diskInfo.UsedPercent = info.UsedPercent
		diskInfoList = append(diskInfoList, diskInfo)
	}
	return
}

// GetNetworkInfo retrieves network information.
//
// No parameters.
// Returns a slice of repo.Network.
func GetNetworkInfo() (networkList []repo.Network) {
	netIOs, _ := net.IOCounters(true)
	if netIOs == nil {
		return
	}
	for _, netIO := range netIOs {
		network := repo.Network{}
		network.Name = netIO.Name
		network.BytesSent = netIO.BytesSent
		network.BytesRecv = netIO.BytesRecv
		network.PacketsSent = netIO.PacketsSent
		network.PacketsRecv = netIO.PacketsRecv
		networkList = append(networkList, network)
	}
	return
}
