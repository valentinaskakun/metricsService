package main

import (
	"testing"
)

//todo: дописать тесты для sendMetrics/sendPost (?)
func TestUpdateGaugeMetrics(t *testing.T) {
	if updateGaugeMetrics()["randomValue"] == updateGaugeMetrics()["randomValue"] {
		t.Errorf("randomValue are equal, does it work?")
	}
}
func TestUpdateCounterMetrics(t *testing.T) {
	test := map[string]int64{"pollCount": 6}
	if updateCounterMetrics("add", test)["pollCount"] != 7 {
		t.Errorf("pollCount didn't incr")
	}
	if updateCounterMetrics("init", test)["pollCount"] != 0 {
		t.Errorf("pollCount didn't init")
	}
}
