package systemmetrics

import (
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
)

type SystemMetrics struct {
	CPUPercent []float64 `json:"cpu_percent"`
	RAM        RAMInfo   `json:"ram"`
	Disk       DiskInfo  `json:"disk"`
	Network    NetInfo   `json:"network"`
	Timestamp  time.Time `json:"timestamp"`
}

type RAMInfo struct {
	Total uint64  `json:"total_gb"`
	Used  uint64  `json:"used_gb"`
	Free  uint64  `json:"free_gb"`
	Usage float64 `json:"usage_percent"`
}

type DiskInfo struct {
	Total uint64  `json:"total_gb"`
	Used  uint64  `json:"used_gb"`
	Free  uint64  `json:"free_gb"`
	Usage float64 `json:"usage_percent"`
}

type NetInfo struct {
	BytesSent uint64 `json:"bytes_sent"`
	BytesRecv uint64 `json:"bytes_recv"`
}

func GetSystemMetrics() (*SystemMetrics, error) {
	metrics := &SystemMetrics{Timestamp: time.Now()}

	// CPU usage (%)
	cpuPercents, err := cpu.Percent(time.Second, false)
	if err != nil {
		return nil, err
	}
	metrics.CPUPercent = cpuPercents

	// RAM info
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}
	metrics.RAM = RAMInfo{
		Total: memInfo.Total / 1024 / 1024 / 1024,
		Used:  memInfo.Used / 1024 / 1024 / 1024,
		Free:  memInfo.Free / 1024 / 1024 / 1024,
		Usage: memInfo.UsedPercent,
	}

	// Disk info
	diskInfo, err := disk.Usage("/")
	if err != nil {
		return nil, err
	}
	metrics.Disk = DiskInfo{
		Total: diskInfo.Total / 1024 / 1024 / 1024,
		Used:  diskInfo.Used / 1024 / 1024 / 1024,
		Free:  diskInfo.Free / 1024 / 1024 / 1024,
		Usage: diskInfo.UsedPercent,
	}

	// Network IO (total desde arranque)
	netIO, err := net.IOCounters(false)
	if err != nil {
		return nil, err
	}
	if len(netIO) > 0 {
		metrics.Network = NetInfo{
			BytesSent: netIO[0].BytesSent,
			BytesRecv: netIO[0].BytesRecv,
		}
	}

	return metrics, nil
}
