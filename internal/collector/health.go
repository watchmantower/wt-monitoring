package collector

import (
	"context"
	"net"
	"os/exec"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"server-monitor/internal/model"

	"github.com/shirou/gopsutil/v3/process"
)

const (
	collectorCommandTimeout = 2 * time.Second
	portCheckTimeout        = 250 * time.Millisecond
	localhostAddress        = "127.0.0.1"
)

type serviceTarget struct {
	Name         string
	SystemdNames []string
	ProcessNames []string
}

var monitoredServices = []serviceTarget{
	{Name: "nginx", SystemdNames: []string{"nginx"}, ProcessNames: []string{"nginx"}},
	{Name: "docker", SystemdNames: []string{"docker", "docker.service"}, ProcessNames: []string{"dockerd", "docker"}},
	{Name: "mongod", SystemdNames: []string{"mongod", "mongodb", "mongodb.service"}, ProcessNames: []string{"mongod"}},
	{Name: "redis", SystemdNames: []string{"redis", "redis-server", "redis.service"}, ProcessNames: []string{"redis-server", "redis"}},
}
var monitoredPorts = []int{80, 443, 27017, 6379}

func CollectSupplementalMetrics() model.SupplementalMetrics {
	serviceHealth := collectServiceHealth(monitoredServices)
	portHealth := collectPortHealth(monitoredPorts)
	nginxSites := collectNginxSites(serviceHealth)

	return model.SupplementalMetrics{
		ServiceHealth: serviceHealth,
		PortHealth:    portHealth,
		DockerHealth:  collectDockerHealth(serviceHealth),
		NginxHealth:   collectNginxHealth(serviceHealth),
		NginxSites:    nginxSites,
		MongoHealth:   collectMongoHealth(serviceHealth, portHealth),
		RedisHealth:   collectRedisHealth(serviceHealth, portHealth),
	}
}

func collectServiceHealth(targets []serviceTarget) []model.ServiceHealth {
	health := make([]model.ServiceHealth, 0, len(targets))
	for _, target := range targets {
		status, source := detectServiceStatus(target)
		health = append(health, model.ServiceHealth{
			Name:   target.Name,
			Status: status,
			Source: source,
		})
	}

	return health
}

func collectPortHealth(ports []int) []model.PortHealth {
	health := make([]model.PortHealth, 0, len(ports))
	for _, port := range ports {
		health = append(health, model.PortHealth{
			Port:      port,
			Listening: isLocalTCPPortListening(port),
			Protocol:  "tcp",
		})
	}

	return health
}

func collectDockerHealth(services []model.ServiceHealth) *model.DockerHealth {
	if findServiceStatus(services, "docker") != "active" {
		return &model.DockerHealth{Available: false}
	}

	statuses, ok := collectDockerContainerStatuses()
	if !ok {
		return &model.DockerHealth{Available: false}
	}

	health := &model.DockerHealth{Available: true}
	for _, status := range statuses {
		lowerStatus := strings.ToLower(status)
		switch {
		case strings.HasPrefix(lowerStatus, "restarting"):
			health.RestartingContainers++
		case strings.Contains(lowerStatus, "unhealthy"):
			health.UnhealthyContainers++
			health.RunningContainers++
		case strings.HasPrefix(lowerStatus, "up"):
			health.RunningContainers++
		default:
			health.StoppedContainers++
		}
	}

	return health
}

func collectNginxHealth(services []model.ServiceHealth) *model.NginxHealth {
	status := findServiceStatus(services, "nginx")
	return &model.NginxHealth{
		Available:     status == "active",
		ServiceStatus: status,
	}
}

func collectNginxSites(services []model.ServiceHealth) []model.NginxSite {
	if findServiceStatus(services, "nginx") != "active" {
		return nil
	}

	output, ok := runCommand("nginx", "-T")
	if !ok {
		return nil
	}

	return parseNginxConfig(output)
}

func collectMongoHealth(services []model.ServiceHealth, ports []model.PortHealth) *model.MongoHealth {
	serviceStatus := findServiceStatus(services, "mongod")
	portListening := findPortListening(ports, 27017)

	return &model.MongoHealth{
		Available:     serviceStatus == "active" || portListening,
		ServiceStatus: serviceStatus,
		PortListening: portListening,
	}
}

func collectRedisHealth(services []model.ServiceHealth, ports []model.PortHealth) *model.RedisHealth {
	serviceStatus := findServiceStatus(services, "redis")
	portListening := findPortListening(ports, 6379)

	return &model.RedisHealth{
		Available:     serviceStatus == "active" || portListening,
		ServiceStatus: serviceStatus,
		PortListening: portListening,
	}
}

func detectServiceStatus(target serviceTarget) (string, string) {
	if status := detectSystemdServiceStatus(target.SystemdNames); status != "" && status != "unknown" {
		return status, "systemd"
	}

	if detectProcessRunning(target.ProcessNames) {
		return "active", "process"
	}

	return "unknown", "none"
}

func detectSystemdServiceStatus(serviceNames []string) string {
	for _, serviceName := range serviceNames {
		ctx, cancel := context.WithTimeout(context.Background(), collectorCommandTimeout)
		output, _ := exec.CommandContext(ctx, "systemctl", "is-active", serviceName).CombinedOutput()
		cancel()

		status := normalizeSystemdStatus(string(output))
		if status != "" && status != "unknown" {
			return status
		}
	}

	return "unknown"
}

func normalizeSystemdStatus(raw string) string {
	status := strings.TrimSpace(raw)
	switch status {
	case "active", "inactive", "failed", "activating", "deactivating", "reloading":
		return status
	default:
		return "unknown"
	}
}

func isLocalTCPPortListening(port int) bool {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(localhostAddress, strconv.Itoa(port)), portCheckTimeout)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func collectDockerContainerStatuses() ([]string, bool) {
	output, ok := runCommand("docker", "ps", "-a", "--format", "{{.Status}}")
	if !ok {
		return nil, false
	}

	trimmed := strings.TrimSpace(output)
	if trimmed == "" {
		return []string{}, true
	}

	return strings.Split(trimmed, "\n"), true
}

func runCommand(name string, args ...string) (string, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), collectorCommandTimeout)
	defer cancel()

	output, err := exec.CommandContext(ctx, name, args...).CombinedOutput()
	if err != nil {
		return "", false
	}

	return string(output), true
}

func parseNginxConfig(raw string) []model.NginxSite {
	const filePrefix = "# configuration file "

	lines := strings.Split(raw, "\n")
	sites := []model.NginxSite{}
	currentFile := ""
	braceDepth := 0
	serverDepth := -1
	var currentSite *model.NginxSite

	for _, rawLine := range lines {
		trimmedRaw := strings.TrimSpace(rawLine)
		if strings.HasPrefix(trimmedRaw, filePrefix) {
			currentFile = strings.TrimSuffix(strings.TrimPrefix(trimmedRaw, filePrefix), ":")
		}

		trimmed := strings.TrimSpace(stripInlineComment(rawLine))
		if trimmed == "" {
			braceDepth += strings.Count(rawLine, "{")
			braceDepth -= strings.Count(rawLine, "}")
			continue
		}

		if isServerBlockStart(trimmed) {
			serverDepth = braceDepth + strings.Count(trimmed, "{")
			currentSite = &model.NginxSite{
				ConfigPath: currentFile,
			}
		} else if currentSite != nil {
			switch {
			case strings.HasPrefix(trimmed, "server_name "):
				currentSite.ServerNames = appendUniqueStrings(currentSite.ServerNames, parseNginxDirectiveValues(trimmed)...)
			case strings.HasPrefix(trimmed, "root "):
				values := parseNginxDirectiveValues(trimmed)
				if len(values) > 0 {
					currentSite.Root = values[0]
				}
			case strings.HasPrefix(trimmed, "listen "):
				currentSite.ListenPorts = appendUniqueInts(currentSite.ListenPorts, parseListenPorts(trimmed)...)
			}
		}

		braceDepth += strings.Count(rawLine, "{")
		braceDepth -= strings.Count(rawLine, "}")

		if currentSite != nil && braceDepth < serverDepth {
			if isMeaningfulNginxSite(*currentSite) {
				sites = append(sites, *currentSite)
			}
			currentSite = nil
			serverDepth = -1
		}
	}

	return sites
}

func isServerBlockStart(line string) bool {
	return line == "server {" || strings.HasPrefix(line, "server{")
}

func stripInlineComment(line string) string {
	commentIndex := strings.Index(line, "#")
	if commentIndex == -1 {
		return line
	}

	return line[:commentIndex]
}

func parseNginxDirectiveValues(line string) []string {
	withoutSemicolon := strings.TrimSuffix(strings.TrimSpace(line), ";")
	parts := strings.Fields(withoutSemicolon)
	if len(parts) <= 1 {
		return nil
	}

	values := []string{}
	for _, part := range parts[1:] {
		if part == "_" {
			continue
		}
		values = append(values, part)
	}

	return values
}

func parseListenPorts(line string) []int {
	values := parseNginxDirectiveValues(line)
	ports := []int{}
	portPattern := regexp.MustCompile(`(?::|\[::\]:)?(\d{2,5})$`)

	for _, value := range values {
		if value == "default_server" || value == "ssl" || value == "http2" || value == "proxy_protocol" {
			continue
		}

		match := portPattern.FindStringSubmatch(value)
		if len(match) < 2 {
			if port, err := strconv.Atoi(value); err == nil {
				ports = appendUniqueInts(ports, port)
			}
			continue
		}

		if port, err := strconv.Atoi(match[1]); err == nil {
			ports = appendUniqueInts(ports, port)
		}
	}

	return ports
}

func appendUniqueStrings(existing []string, incoming ...string) []string {
	for _, item := range incoming {
		if item == "" || slices.Contains(existing, item) {
			continue
		}
		existing = append(existing, item)
	}

	return existing
}

func appendUniqueInts(existing []int, incoming ...int) []int {
	for _, item := range incoming {
		if item == 0 || slices.Contains(existing, item) {
			continue
		}
		existing = append(existing, item)
	}

	return existing
}

func isMeaningfulNginxSite(site model.NginxSite) bool {
	return len(site.ServerNames) > 0 || site.Root != "" || len(site.ListenPorts) > 0 || site.ConfigPath != ""
}

func detectProcessRunning(processNames []string) bool {
	processes, err := process.Processes()
	if err != nil {
		return false
	}

	for _, proc := range processes {
		name, err := proc.Name()
		if err != nil {
			continue
		}

		lowerName := strings.ToLower(name)
		for _, processName := range processNames {
			if lowerName == strings.ToLower(processName) {
				return true
			}
		}
	}

	return false
}

func findServiceStatus(services []model.ServiceHealth, serviceName string) string {
	for _, service := range services {
		if service.Name == serviceName {
			return service.Status
		}
	}

	return "unknown"
}

func findPortListening(ports []model.PortHealth, port int) bool {
	for _, item := range ports {
		if item.Port == port {
			return item.Listening
		}
	}

	return false
}
