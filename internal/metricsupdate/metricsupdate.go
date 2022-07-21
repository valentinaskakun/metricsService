package metricsupdate

import (
	"math/rand"

	"github.com/valentinaskakun/metricsService/internal/metricscollect"
	"github.com/valentinaskakun/metricsService/internal/storage"
)

func UpdateGaugeMetricsRuntime(metricsRun *storage.Metrics) {
	tempCurrentMemStatsMetrics := metricscollect.GetCurrentValuesRuntimeGauge()
	metricsRun.MuGauge.Lock()
	for key, value := range tempCurrentMemStatsMetrics {
		metricsRun.GaugeMetric[key] = value
	}
	metricsRun.MuGauge.Unlock()
	metricsRun.GaugeMetric["RandomValue"] = rand.Float64()
}
func UpdateGaugeMetricsCPU(metricsRun *storage.Metrics) {
	tempCurrentCPUStatsMetrics := metricscollect.GetCurrentValuesGOpsGauge()
	metricsRun.MuGauge.Lock()
	for key, value := range tempCurrentCPUStatsMetrics {
		metricsRun.GaugeMetric[key] = value
	}
	metricsRun.MuGauge.Unlock()
}
func UpdateCounterMetrics(action string, metricsRun *storage.Metrics) {
	switch {
	case action == "add":
		metricsRun.MuCounter.Lock()
		defer metricsRun.MuCounter.Unlock()
		for key, value := range metricsRun.CounterMetric {
			metricsRun.CounterMetric[key] = value + 1
		}
	case action == "init":
		metricsRun.MuCounter.Lock()
		defer metricsRun.MuCounter.Unlock()
		for key := range metricsRun.CounterMetric {
			metricsRun.CounterMetric[key] = 0
		}
	}
	return
}
