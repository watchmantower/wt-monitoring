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

type APIResponse struct {
    Status   string `json:"status"`
    Interval int    `json:"interval"`
    Message  string `json:"message"`
}

func collectMetrics(serverID string) Metrics {
    cpuPercentages, _ := cpu.Percent(0, false)
    cpuUsage := cpuPercentages[0]
    vmStats, _ := mem.VirtualMemory()
    memoryUsage := vmStats.UsedPercent
    totalMemory := vmStats.Total / 1024 / 1024
    usedMemory := vmStats.Used / 1024 / 1024
    diskStats, _ := disk.Usage("/")
    diskUsage := diskStats.UsedPercent
    totalDisk := diskStats.Total / 1024 / 1024
    usedDisk := diskStats.Used / 1024 / 1024
    netIOStats, _ := net.IOCounters(false)
    networkSent := netIOStats[0].BytesSent / 1024 / 1024
    networkReceived := netIOStats[0].BytesRecv / 1024 / 1024
    loadStats, _ := load.Avg()
    load1 := loadStats.Load1
    load5 := loadStats.Load5
    load15 := loadStats.Load15
    uptime, _ := host.Uptime()
    hostInfo, _ := host.Info()
    totalProcesses := hostInfo.Procs
    swapStats, _ := mem.SwapMemory()
    swapTotal := swapStats.Total / 1024 / 1024
    swapUsed := swapStats.Used / 1024 / 1024
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

func sendMetrics(apiURL string, apiKey string, metrics Metrics) (int, error) {
    jsonData, err := json.Marshal(metrics)
    if err != nil {
        return 10, err
    }

    req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
    if err != nil {
        return 10, err
    }
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return 10, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return 10, fmt.Errorf("Sunucu hatası: %s", resp.Status)
    }

    var apiResponse APIResponse
    if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
        return 10, err
    }

    if apiResponse.Interval == 0 {
        fmt.Println("API yanıtında interval değeri eksik, varsayılan interval kullanılacak.")
        apiResponse.Interval = 10
    }

    fmt.Printf("Metrikler başarıyla gönderildi. Yeni interval: %d saniye\n", apiResponse.Interval)

    return apiResponse.Interval, nil
}

func main() {
    serverID := flag.String("server_id", "", "Server ID")
    apiKey := flag.String("api_key", "", "API Key")
    apiURL := flag.String("api_url", "", "WT API URL")
    flag.Parse()

    if *serverID == "" || *apiKey == "" {
        fmt.Println("Server ID ve API Key gerekli")
        return
    }

    interval := 10 // Varsayılan bekleme süresi

    for {
        metrics := collectMetrics(*serverID)

        newInterval, err := sendMetrics(*apiURL, *apiKey, metrics)
        if err != nil {
            fmt.Println("Metrik gönderim hatası:", err)
            // Hata durumunda eski interval süresini koru
        } else {
            interval = newInterval
        }

        time.Sleep(time.Duration(interval) * time.Second)
    }
}