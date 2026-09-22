package monitor

import (
	"fmt"
	"strings"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

type DiskStat struct {
	Device      string
	Mountpoint  string
	UsedPercent float64
}

type Stats struct {
	CPUPercent float64
	RAMPercent float64
	RAMUsedGB  float64
	RAMTotalGB float64
	Disks      []DiskStat
}

func Collect(diskDevices []string, rootPrefix string) (*Stats, error) {
	cpuPercents, err := cpu.Percent(0, false)
	if err != nil {
		return nil, fmt.Errorf("cpu: %w", err)
	}
	cpuPct := 0.0
	if len(cpuPercents) > 0 {
		cpuPct = cpuPercents[0]
	}

	vm, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("mem: %w", err)
	}

	disks, err := collectDisks(diskDevices, rootPrefix)
	if err != nil {
		return nil, fmt.Errorf("disk: %w", err)
	}

	return &Stats{
		CPUPercent: cpuPct,
		RAMPercent: vm.UsedPercent,
		RAMUsedGB:  bytesToGB(vm.Used),
		RAMTotalGB: bytesToGB(vm.Total),
		Disks:      disks,
	}, nil
}

func collectDisks(devices []string, rootPrefix string) ([]DiskStat, error) {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return nil, err
	}

	seen := map[string]bool{}
	var results []DiskStat

	for _, dev := range devices {
		for _, p := range partitions {
			if !strings.Contains(p.Device, dev) {
				continue
			}
			if seen[p.Mountpoint] {
				continue
			}
			usage, err := disk.Usage(rootPrefix + p.Mountpoint)
			if err != nil {
				continue
			}
			seen[p.Mountpoint] = true
			results = append(results, DiskStat{
				Device:      p.Device,
				Mountpoint:  p.Mountpoint,
				UsedPercent: usage.UsedPercent,
			})
		}
	}

	return results, nil
}

func bytesToGB(b uint64) float64 {
	return float64(b) / 1024 / 1024 / 1024
}
