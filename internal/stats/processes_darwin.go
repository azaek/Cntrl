//go:build darwin

package stats

import (
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

func GetProcesses() ([]Process, error) {
	// Use ps to get PID, RSS (memory in KB), CPU time, and Command
	cmd := exec.Command("ps", "-A", "-o", "pid,rss,time,comm")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(out), "\n")

	// Aggregate by process name
	aggregated := make(map[string]*Process)

	// Skip header
	startIdx := 1
	if len(lines) > 0 && strings.Contains(lines[0], "PID") {
		startIdx = 1
	}

	for i := startIdx; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		// RSS (KB) -> Bytes
		rssKB, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue
		}
		memBytes := rssKB * 1024

		// CPU time
		cpuTime := parseCpuTime(fields[2])

		// Command (Name)
		name := strings.Join(fields[3:], " ")
		if idx := strings.LastIndex(name, "/"); idx != -1 {
			name = name[idx+1:]
		}

		if existing, ok := aggregated[name]; ok {
			existing.Count++
			existing.Memory += memBytes
			existing.CpuTime += cpuTime
		} else {
			aggregated[name] = &Process{
				Name:    name,
				Count:   1,
				Memory:  memBytes,
				CpuTime: cpuTime,
			}
		}
	}

	// Convert map to slice
	results := make([]Process, 0, len(aggregated))
	for _, p := range aggregated {
		p.MemoryMB = float64(p.Memory) / (1024 * 1024)
		results = append(results, *p)
	}

	// Sort by memory (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Memory > results[j].Memory
	})

	// Return top 20
	if len(results) > 20 {
		results = results[:20]
	}

	return results, nil
}

// parseCpuTime parses CPU time string (MM:SS.ss or HH:MM:SS) to seconds
func parseCpuTime(s string) float64 {
	parts := strings.Split(s, ":")
	if len(parts) == 2 {
		minutes, _ := strconv.ParseFloat(parts[0], 64)
		seconds, _ := strconv.ParseFloat(parts[1], 64)
		return minutes*60 + seconds
	} else if len(parts) == 3 {
		hours, _ := strconv.ParseFloat(parts[0], 64)
		minutes, _ := strconv.ParseFloat(parts[1], 64)
		seconds, _ := strconv.ParseFloat(parts[2], 64)
		return hours*3600 + minutes*60 + seconds
	}
	return 0
}
