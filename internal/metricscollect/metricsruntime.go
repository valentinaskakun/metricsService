package metricscollect

import (
	"runtime"
)

//todo: добавить обработку ошибок(?)
func GetCurrentValuesRuntimeGauge() map[string]float64 {
	currentValuesRuntimeGauge := make(map[string]float64)
	currentMemStats := new(runtime.MemStats)
	runtime.ReadMemStats(currentMemStats)
	currentValuesRuntimeGauge["Alloc"] = float64(currentMemStats.Alloc)
	currentValuesRuntimeGauge["BuckHashSys"] = float64(currentMemStats.BuckHashSys)
	currentValuesRuntimeGauge["Frees"] = float64(currentMemStats.Frees)
	currentValuesRuntimeGauge["GCCPUFraction"] = float64(currentMemStats.GCCPUFraction)
	currentValuesRuntimeGauge["GCSys"] = float64(currentMemStats.GCSys)
	currentValuesRuntimeGauge["HeapAlloc"] = float64(currentMemStats.HeapAlloc)
	currentValuesRuntimeGauge["HeapIdle"] = float64(currentMemStats.HeapIdle)
	currentValuesRuntimeGauge["HeapInuse"] = float64(currentMemStats.HeapInuse)
	currentValuesRuntimeGauge["HeapObjects"] = float64(currentMemStats.HeapObjects)
	currentValuesRuntimeGauge["HeapReleased"] = float64(currentMemStats.HeapReleased)
	currentValuesRuntimeGauge["HeapSys"] = float64(currentMemStats.HeapSys)
	currentValuesRuntimeGauge["LastGC"] = float64(currentMemStats.LastGC)
	currentValuesRuntimeGauge["Lookups"] = float64(currentMemStats.Lookups)
	currentValuesRuntimeGauge["MCacheInuse"] = float64(currentMemStats.MCacheInuse)
	currentValuesRuntimeGauge["MCacheSys"] = float64(currentMemStats.MCacheSys)
	currentValuesRuntimeGauge["MSpanInuse"] = float64(currentMemStats.MSpanInuse)
	currentValuesRuntimeGauge["MSpanSys"] = float64(currentMemStats.MSpanSys)
	currentValuesRuntimeGauge["Mallocs"] = float64(currentMemStats.Mallocs)
	currentValuesRuntimeGauge["NextGC"] = float64(currentMemStats.NextGC)
	currentValuesRuntimeGauge["NumForcedGC"] = float64(currentMemStats.NumForcedGC)
	currentValuesRuntimeGauge["NumGC"] = float64(currentMemStats.NumGC)
	currentValuesRuntimeGauge["OtherSys"] = float64(currentMemStats.OtherSys)
	currentValuesRuntimeGauge["PauseTotalNs"] = float64(currentMemStats.PauseTotalNs)
	currentValuesRuntimeGauge["StackInuse"] = float64(currentMemStats.StackInuse)
	currentValuesRuntimeGauge["StackSys"] = float64(currentMemStats.StackSys)
	currentValuesRuntimeGauge["Sys"] = float64(currentMemStats.Sys)
	currentValuesRuntimeGauge["TotalAlloc"] = float64(currentMemStats.TotalAlloc)
	return currentValuesRuntimeGauge
}
