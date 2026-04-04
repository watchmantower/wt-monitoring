package model

const AppVersion = "1.4.0"

type AgentInfo struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	Platform string `json:"platform"`
	Arch     string `json:"arch"`
}

type ProcessMetrics struct {
	PID         int32   `json:"pid"`
	Name        string  `json:"name"`
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage uint64  `json:"memory_usage"`
}

type ServiceHealth struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Source string `json:"source"`
}

type PortHealth struct {
	Port      int    `json:"port"`
	Listening bool   `json:"listening"`
	Protocol  string `json:"protocol"`
}

type DockerHealth struct {
	Available            bool `json:"available"`
	RunningContainers    int  `json:"running_containers"`
	StoppedContainers    int  `json:"stopped_containers"`
	RestartingContainers int  `json:"restarting_containers"`
	UnhealthyContainers  int  `json:"unhealthy_containers"`
}

type NginxHealth struct {
	Available     bool   `json:"available"`
	ServiceStatus string `json:"service_status"`
}

type NginxSite struct {
	ServerNames []string `json:"server_names,omitempty"`
	Root        string   `json:"root,omitempty"`
	ConfigPath  string   `json:"config_path,omitempty"`
	ListenPorts []int    `json:"listen_ports,omitempty"`
}

type MongoHealth struct {
	Available     bool   `json:"available"`
	ServiceStatus string `json:"service_status"`
	PortListening bool   `json:"port_listening"`
}

type RedisHealth struct {
	Available     bool   `json:"available"`
	ServiceStatus string `json:"service_status"`
	PortListening bool   `json:"port_listening"`
}

type SupplementalMetrics struct {
	ServiceHealth []ServiceHealth `json:"service_health,omitempty"`
	PortHealth    []PortHealth    `json:"port_health,omitempty"`
	DockerHealth  *DockerHealth   `json:"docker_health,omitempty"`
	NginxHealth   *NginxHealth    `json:"nginx_health,omitempty"`
	NginxSites    []NginxSite     `json:"nginx_sites,omitempty"`
	MongoHealth   *MongoHealth    `json:"mongo_health,omitempty"`
	RedisHealth   *RedisHealth    `json:"redis_health,omitempty"`
}

type Metrics struct {
	ServerID        string           `json:"server_id"`
	AppVersion      string           `json:"app_version"`
	Agent           AgentInfo        `json:"agent"`
	CPUUsage        float64          `json:"cpu_usage"`
	MemoryUsage     float64          `json:"memory_usage"`
	TotalMemory     uint64           `json:"total_memory"`
	UsedMemory      uint64           `json:"used_memory"`
	DiskUsage       float64          `json:"disk_usage"`
	TotalDisk       uint64           `json:"total_disk"`
	UsedDisk        uint64           `json:"used_disk"`
	NetworkSent     uint64           `json:"network_sent"`
	NetworkReceived uint64           `json:"network_received"`
	Load1           float64          `json:"load_1"`
	Load5           float64          `json:"load_5"`
	Load15          float64          `json:"load_15"`
	Uptime          uint64           `json:"uptime"`
	SwapTotal       uint64           `json:"swap_total"`
	SwapUsed        uint64           `json:"swap_used"`
	SwapUsage       float64          `json:"swap_usage"`
	OS              string           `json:"os,omitempty"`
	Platform        string           `json:"platform,omitempty"`
	PlatformVersion string           `json:"platform_version,omitempty"`
	KernelVersion   string           `json:"kernel_version,omitempty"`
	Processes       []ProcessMetrics `json:"processes"`
	ServiceHealth   []ServiceHealth  `json:"service_health,omitempty"`
	PortHealth      []PortHealth     `json:"port_health,omitempty"`
	DockerHealth    *DockerHealth    `json:"docker_health,omitempty"`
	NginxHealth     *NginxHealth     `json:"nginx_health,omitempty"`
	NginxSites      []NginxSite      `json:"nginx_sites,omitempty"`
	MongoHealth     *MongoHealth     `json:"mongo_health,omitempty"`
	RedisHealth     *RedisHealth     `json:"redis_health,omitempty"`
}

type APIResponse struct {
	Status   string `json:"status"`
	Interval int    `json:"interval"`
	Message  string `json:"message"`
}
