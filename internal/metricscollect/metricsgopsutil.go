package metricscollect

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
)

func GetCurrentValuesGOpsGauge() map[string]float64 {
	currentValuesRuntimeGauge := make(map[string]float64)
	currentMemStats, _ := mem.VirtualMemory()
	currentCPUStats, _ := cpu.Percent(3*time.Second, true)
	currentCPUload, _ := load.Avg()
	fmt.Println("cpuload", currentCPUload)
	fmt.Println(currentCPUStats)
	//runtime.ReadMemStats(currentMemStats)
	currentValuesRuntimeGauge["TotalMemory"] = float64(currentMemStats.Total)
	currentValuesRuntimeGauge["FreeMemory"] = float64(currentMemStats.Free)
	//сделать правильно
	currentValuesRuntimeGauge["CPUutilization1"] = rand.Float64()
	return currentValuesRuntimeGauge
}
