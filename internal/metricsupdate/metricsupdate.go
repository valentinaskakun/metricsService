package metricsupdate

import (
	"math/rand"

	"github.com/valentinaskakun/metricsService/internal/metricscollect"
	"github.com/valentinaskakun/metricsService/internal/storage"
)

func UpdateGaugeMetricsRuntime(metricsRun *storage.Metrics) {
	tempCurrentMemStatsMetrics := metricscollect.GetCurrentValuesRuntimeGauge()
	for key, value := range tempCurrentMemStatsMetrics {
		metricsRun.GaugeMetric[key] = value
	}
	metricsRun.GaugeMetric["RandomValue"] = rand.Float64()
}
func UpdateGaugeMetricsCPU(metricsRun *storage.Metrics) {
	tempCurrentCPUStatsMetrics := metricscollect.GetCurrentValuesGOpsGauge()
	for key, value := range tempCurrentCPUStatsMetrics {
		metricsRun.GaugeMetric[key] = value
	}
}
func UpdateCounterMetrics(action string, metricsCounterToUpdate map[string]int64) (metricsCounterUpdated map[string]int64) {
	metricsCounterUpdated = make(map[string]int64)
	if _, ok := metricsCounterToUpdate["PollCount"]; !ok {
		metricsCounterToUpdate = make(map[string]int64)
		metricsCounterToUpdate["PollCount"] = 0
	}
	switch {
	case action == "add":
		for key, value := range metricsCounterToUpdate {
			metricsCounterUpdated[key] = value + 1
		}
	case action == "init":
		for key := range metricsCounterToUpdate {
			metricsCounterUpdated[key] = 0
		}
	}
	return metricsCounterUpdated
}
