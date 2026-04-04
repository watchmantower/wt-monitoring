package collector

import (
	"runtime"
	"server-monitor/internal/model"

	"github.com/shirou/gopsutil/v3/host"
)

func CollectStaticMetrics(serverID string) model.Metrics {
	hostInfo, _ := host.Info()

	return model.Metrics{
		ServerID:   serverID,
		AppVersion: model.AppVersion,
		Agent: model.AgentInfo{
			Name:     "warden",
			Version:  model.AppVersion,
			Platform: runtime.GOOS,
			Arch:     runtime.GOARCH,
		},
		OS:              hostInfo.OS,
		Platform:        hostInfo.Platform,
		PlatformVersion: hostInfo.PlatformVersion,
		KernelVersion:   hostInfo.KernelVersion,
	}
}
