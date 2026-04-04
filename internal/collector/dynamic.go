package collector

import (
	"server-monitor/internal/model"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

const (
	bytesToMB    = 1024 * 1024
	rootDiskPath = "/"
)

func CollectDynamicMetrics(processes []model.ProcessMetrics, supplemental model.SupplementalMetrics) model.Metrics {
	cpuPercentages, _ := cpu.Percent(0, false)
	cpuUsage := 0.0
	if len(cpuPercentages) > 0 {
		cpuUsage = cpuPercentages[0]
	}

	vmStats, _ := mem.VirtualMemory()
	memoryUsage := vmStats.UsedPercent
	totalMemory := vmStats.Total / bytesToMB
	usedMemory := vmStats.Used / bytesToMB

	diskStats, _ := disk.Usage(rootDiskPath)
	diskUsage := diskStats.UsedPercent
	totalDisk := diskStats.Total / bytesToMB
	usedDisk := diskStats.Used / bytesToMB

	netIOStats, _ := net.IOCounters(false)
	var networkSent uint64
	var networkReceived uint64
	if len(netIOStats) > 0 {
		networkSent = netIOStats[0].BytesSent / bytesToMB
		networkReceived = netIOStats[0].BytesRecv / bytesToMB
	}

	loadStats, _ := load.Avg()
	load1 := loadStats.Load1
	load5 := loadStats.Load5
	load15 := loadStats.Load15
	uptime, _ := host.Uptime()
	swapStats, _ := mem.SwapMemory()
	swapTotal := swapStats.Total / bytesToMB
	swapUsed := swapStats.Used / bytesToMB
	swapUsage := swapStats.UsedPercent

	return model.Metrics{
		CPUUsage:        cpuUsage,
		MemoryUsage:     memoryUsage,
		TotalMemory:     totalMemory,
		UsedMemory:      usedMemory,
		DiskUsage:       diskUsage,
		TotalDisk:       totalDisk,
		UsedDisk:        usedDisk,
		NetworkSent:     networkSent,
		NetworkReceived: networkReceived,
		Load1:           load1,
		Load5:           load5,
		Load15:          load15,
		Uptime:          uptime,
		SwapTotal:       swapTotal,
		SwapUsed:        swapUsed,
		SwapUsage:       swapUsage,
		Processes:       processes,
		ServiceHealth:   supplemental.ServiceHealth,
		PortHealth:      supplemental.PortHealth,
		DockerHealth:    supplemental.DockerHealth,
		NginxHealth:     supplemental.NginxHealth,
		NginxSites:      supplemental.NginxSites,
		MongoHealth:     supplemental.MongoHealth,
		RedisHealth:     supplemental.RedisHealth,
	}
}
