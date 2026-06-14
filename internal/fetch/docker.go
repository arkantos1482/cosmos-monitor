package fetch

import (
	"fmt"
	"sync"
	"time"
)

// cpuCache holds the previous stats snapshot for computing CPU delta across calls.
type cpuCacheEntry struct {
	cpuTotal uint64
	cpuSys   uint64
}

var (
	cpuCacheMu sync.Mutex
	cpuCache   = map[string]cpuCacheEntry{}
)

// DockerSnapshot holds Docker container metrics.
type DockerSnapshot struct {
	Running      bool
	Image        string
	OOMKilled    bool
	CPUPercent   float64
	MemUsage     uint64
	MemLimit     uint64
	RestartCount int
	StartedAt    time.Time
	Err          error
}

type dockerStats struct {
	CPUStats struct {
		CPUUsage struct {
			TotalUsage uint64 `json:"total_usage"`
		} `json:"cpu_usage"`
		SystemCPUUsage uint64 `json:"system_cpu_usage"`
		OnlineCPUs     int    `json:"online_cpus"`
	} `json:"cpu_stats"`
	PreCPUStats struct {
		CPUUsage struct {
			TotalUsage uint64 `json:"total_usage"`
		} `json:"cpu_usage"`
		SystemCPUUsage uint64 `json:"system_cpu_usage"`
	} `json:"precpu_stats"`
	MemoryStats struct {
		Usage uint64 `json:"usage"`
		Limit uint64 `json:"limit"`
	} `json:"memory_stats"`
}

type dockerInspect struct {
	Config struct {
		Image string `json:"Image"`
	} `json:"Config"`
	State struct {
		Running      bool   `json:"Running"`
		StartedAt    string `json:"StartedAt"`
		Status       string `json:"Status"`
		OOMKilled    bool   `json:"OOMKilled"`
		RestartCount int    `json:"RestartCount"`
	} `json:"State"`
	RestartCount int `json:"RestartCount"`
}

// FetchDocker fetches Docker container stats and inspect data.
func FetchDocker(container string) DockerSnapshot {
	client := newDockerClient()
	snap := DockerSnapshot{}

	var insp dockerInspect
	if err := doDockerJSON(client, fmt.Sprintf("http://localhost/containers/%s/json", container), &insp); err != nil {
		snap.Err = err
		return snap
	}

	snap.Running = insp.State.Running
	snap.Image = insp.Config.Image
	snap.OOMKilled = insp.State.OOMKilled
	snap.RestartCount = insp.RestartCount
	if snap.RestartCount == 0 {
		snap.RestartCount = insp.State.RestartCount
	}

	if t, err := time.Parse(time.RFC3339Nano, insp.State.StartedAt); err == nil {
		snap.StartedAt = t
	}

	var stats dockerStats
	if err := doDockerJSON(client, fmt.Sprintf("http://localhost/containers/%s/stats?stream=false", container), &stats); err != nil {
		snap.Err = err
		return snap
	}

	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage)
	sysDelta := float64(stats.CPUStats.SystemCPUUsage - stats.PreCPUStats.SystemCPUUsage)

	// When stream=false on the first call, precpu_stats == cpu_stats → sysDelta == 0.
	// Fall back to the previous snapshot stored across calls.
	cpuCacheMu.Lock()
	prev, hasPrev := cpuCache[container]
	cpuCache[container] = cpuCacheEntry{
		cpuTotal: stats.CPUStats.CPUUsage.TotalUsage,
		cpuSys:   stats.CPUStats.SystemCPUUsage,
	}
	cpuCacheMu.Unlock()

	if sysDelta == 0 && hasPrev {
		cpuDelta = float64(stats.CPUStats.CPUUsage.TotalUsage - prev.cpuTotal)
		sysDelta = float64(stats.CPUStats.SystemCPUUsage - prev.cpuSys)
	}

	numCPUs := stats.CPUStats.OnlineCPUs
	if numCPUs == 0 {
		numCPUs = 1
	}
	if sysDelta > 0 {
		snap.CPUPercent = (cpuDelta / sysDelta) * float64(numCPUs) * 100.0
	}

	snap.MemUsage = stats.MemoryStats.Usage
	snap.MemLimit = stats.MemoryStats.Limit

	return snap
}
