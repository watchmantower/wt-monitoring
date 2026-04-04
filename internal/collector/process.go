package collector

import (
	"sort"

	"server-monitor/internal/model"

	"github.com/shirou/gopsutil/v3/process"
)

const bytesToKB = 1024

func CollectProcessMetrics(maxProcesses int) []model.ProcessMetrics {
	if maxProcesses == 0 {
		return []model.ProcessMetrics{}
	}

	processList := []model.ProcessMetrics{}
	processes, _ := process.Processes()

	for _, proc := range processes {
		name, err := proc.Name()
		if err != nil {
			continue
		}

		cpuPercent, err := proc.CPUPercent()
		if err != nil {
			continue
		}

		memInfo, err := proc.MemoryInfo()
		if err != nil {
			continue
		}

		processList = append(processList, model.ProcessMetrics{
			PID:         proc.Pid,
			Name:        name,
			CPUUsage:    cpuPercent,
			MemoryUsage: memInfo.RSS / bytesToKB,
		})
	}

	sort.Slice(processList, func(i, j int) bool {
		return processList[i].CPUUsage > processList[j].CPUUsage
	})

	if len(processList) > maxProcesses {
		processList = processList[:maxProcesses]
	}

	return processList
}
