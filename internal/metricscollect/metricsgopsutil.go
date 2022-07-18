package metricscollect

import (
	"strconv"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

func GetCurrentValuesGOpsGauge() map[string]float64 {
	currentValuesRuntimeGauge := make(map[string]float64)
	currentMemStats, _ := mem.VirtualMemory()
	currentCPUStats, _ := cpu.Percent(0, true)
	for index, stat := range currentCPUStats {
		val := "CPUutilization" + strconv.Itoa(index+1)
		currentValuesRuntimeGauge[val] = stat
	}
	currentValuesRuntimeGauge["TotalMemory"] = float64(currentMemStats.Total)
	currentValuesRuntimeGauge["FreeMemory"] = float64(currentMemStats.Free)
	//сделать правильно
	return currentValuesRuntimeGauge
}
