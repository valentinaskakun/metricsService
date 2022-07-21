package main

import (
	"testing"

	"github.com/valentinaskakun/metricsService/internal/metricsupdate"
	"github.com/valentinaskakun/metricsService/internal/storage"
)

//todo: дописать тесты для sendMetrics/sendPost (?)

func TestUpdateGaugeMetrics(t *testing.T) {
	var metricsCurrent storage.Metrics
	metricsCurrent.InitMetrics()
	metricsupdate.UpdateGaugeMetricsRuntime(&metricsCurrent)
	val1 := metricsCurrent.GaugeMetric["RandomValue"]
	metricsupdate.UpdateGaugeMetricsRuntime(&metricsCurrent)
	val2 := metricsCurrent.GaugeMetric["RandomValue"]
	if val1 == val2 {
		t.Errorf("RandomValue are equal, does it work?")
	}
}
func TestUpdateCounterMetrics(t *testing.T) {
	var metricsCurrent storage.Metrics
	metricsCurrent.InitMetrics()
	metricsCurrent.CounterMetric["PollCount"] = 6
	metricsupdate.UpdateCounterMetrics("add", &metricsCurrent)
	if metricsCurrent.CounterMetric["PollCount"] != 7 {
		t.Errorf("PollCount didn't incr")
	}
	metricsupdate.UpdateCounterMetrics("init", &metricsCurrent)
	if metricsCurrent.CounterMetric["PollCount"] != 0 {
		t.Errorf("PollCount didn't init")
	}
}
