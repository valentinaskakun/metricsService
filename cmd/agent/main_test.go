package main

import (
	"testing"
)

//todo: дописать тесты для sendMetrics/sendPost (?)
func TestUpdateGaugeMetrics(t *testing.T) {
	if updateGaugeMetricsRuntime()["RandomValue"] == updateGaugeMetricsRuntime()["RandomValue"] {
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
