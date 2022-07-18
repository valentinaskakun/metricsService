package main

import (
	"testing"

	"github.com/valentinaskakun/metricsService/internal/storage"
)

//todo: дописать тесты для sendMetrics/sendPost (?)

func TestUpdateGaugeMetrics(t *testing.T) {
	var metricsCurrent storage.Metrics
	metricsCurrent.InitMetrics()
	updateGaugeMetricsRuntime(&metricsCurrent)
	val1 := metricsCurrent.GaugeMetric["RandomValue"]
	updateGaugeMetricsRuntime(&metricsCurrent)
	val2 := metricsCurrent.GaugeMetric["RandomValue"]
	if val1 == val2 {
		t.Errorf("RandomValue are equal, does it work?")
	}
}
func TestUpdateCounterMetrics(t *testing.T) {
	test := map[string]int64{"PollCount": 6}
	if updateCounterMetrics("add", test)["PollCount"] != 7 {
		t.Errorf("PollCount didn't incr")
	}
	if updateCounterMetrics("init", test)["PollCount"] != 0 {
		t.Errorf("PollCount didn't init")
	}
}
