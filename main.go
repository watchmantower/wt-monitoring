package main

import (
    "bytes"
    "encoding/json"
    "flag"
    "fmt"
    "net/http"
    "time"

    "github.com/shirou/gopsutil/v3/cpu"
    "github.com/shirou/gopsutil/v3/disk"
    "github.com/shirou/gopsutil/v3/host"
    "github.com/shirou/gopsutil/v3/load"
    "github.com/shirou/gopsutil/v3/mem"
    "github.com/shirou/gopsutil/v3/net"
)

type Metrics struct {
    ServerID           string  `json:"server_id"`
    CPUUsage           float64 `json:"cpu_usage"`
    MemoryUsage        float64 `json:"memory_usage"`
    TotalMemory        uint64  `json:"total_memory"`
    UsedMemory         uint64  `json:"used_memory"`
    DiskUsage          float64 `json:"disk_usage"`
    TotalDisk          uint64  `json:"total_disk"`
    UsedDisk           uint64  `json:"used_disk"`
    NetworkSent        uint64  `json:"network_sent"`
    NetworkReceived    uint64  `json:"network_received"`
    Load1              float64 `json:"load_1"`
    Load5              float64 `json:"load_5"`
    Load15             float64 `json:"load_15"`
    Uptime             uint64  `json:"uptime"`
    OpenFileDescriptors uint64 `json:"open_file_descriptors"`
    TotalProcesses     uint64  `json:"total_processes"`
    SwapTotal          uint64  `json:"swap_total"`
    SwapUsed           uint64  `json:"swap_used"`
    SwapUsage          float64 `json:"swap_usage"`
}

func collectMetrics(serverID string) Metrics {
    // CPU kullanımını al
    cpuPercentages, _ := cpu.Percent(0, false)
    cpuUsage := cpuPercentages[0]

    // Bellek kullanımını al
    vmStats, _ := mem.VirtualMemory()
    memoryUsage := vmStats.UsedPercent
    totalMemory := vmStats.Total / 1024 / 1024 // MB
    usedMemory := vmStats.Used / 1024 / 1024   // MB

    // Disk kullanımını al
    diskStats, _ := disk.Usage("/")
    diskUsage := diskStats.UsedPercent
    totalDisk := diskStats.Total / 1024 / 1024 // MB
    usedDisk := diskStats.Used / 1024 / 1024   // MB

    // Ağ (network) kullanımını al
    netIOStats, _ := net.IOCounters(false)
    networkSent := netIOStats[0].BytesSent / 1024 / 1024 // MB
    networkReceived := netIOStats[0].BytesRecv / 1024 / 1024 // MB

    // Load Average (Yük Ortalaması) al
    loadStats, _ := load.Avg()
    load1 := loadStats.Load1
    load5 := loadStats.Load5
    load15 := loadStats.Load15

    // Uptime al
    uptime, _ := host.Uptime()


    // Proses sayısı al
    hostInfo, _ := host.Info()
    totalProcesses := hostInfo.Procs

    // Swap bellek kullanımını al
    swapStats, _ := mem.SwapMemory()
    swapTotal := swapStats.Total / 1024 / 1024 // MB
    swapUsed := swapStats.Used / 1024 / 1024   // MB
    swapUsage := swapStats.UsedPercent

    return Metrics{
        ServerID:           serverID,
        CPUUsage:           cpuUsage,
        MemoryUsage:        memoryUsage,
        TotalMemory:        totalMemory,
        UsedMemory:         usedMemory,
        DiskUsage:          diskUsage,
        TotalDisk:          totalDisk,
        UsedDisk:           usedDisk,
        NetworkSent:        networkSent,
        NetworkReceived:    networkReceived,
        Load1:              load1,
        Load5:              load5,
        Load15:             load15,
        Uptime:             uptime,
        TotalProcesses:     totalProcesses,
        SwapTotal:          swapTotal,
        SwapUsed:           swapUsed,
        SwapUsage:          swapUsage,
    }
}

func sendMetrics(apiURL string, metrics Metrics) error {
    jsonData, err := json.Marshal(metrics)
    if err != nil {
        return err
    }

    resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("Sunucu hatası: %s", resp.Status)
    }

    fmt.Println("Metrikler başarıyla gönderildi:", metrics)
    return nil
}

func main() {
    serverID := flag.String("server_id", "", "Server ID")
    apiURL := flag.String("api_url", "https://wt.com/api/metrics", "WT API URL")
    flag.Parse()

    if *serverID == "" {
        fmt.Println("Server ID gerekli")
        return
    }

    for {
        metrics := collectMetrics(*serverID)

        if err := sendMetrics(*apiURL, metrics); err != nil {
            fmt.Println("Metrik gönderim hatası:", err)
        }

        time.Sleep(10 * time.Second)
    }
}