package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/valentinaskakun/metricsService/internal/config"
	"github.com/valentinaskakun/metricsService/internal/metricssend"
	"github.com/valentinaskakun/metricsService/internal/metricsupdate"
	"github.com/valentinaskakun/metricsService/internal/storage"
)

//todo: навести порядок
const (
	pollIntervalConst   = 2000
	reportIntervalConst = 4000
)

var MetricsCurrent storage.Metrics

func handleSignal(signal os.Signal) {
	fmt.Println("* Got:", signal)
	os.Exit(-1)
}
func main() {
	configRun, err := config.LoadConfigAgent()
	if err != nil {
		log.Println(err)
	}
	pollInterval, err := time.ParseDuration(configRun.PollInterval)
	if err != nil {
		log.Println(err)
	}
	reportInterval, err := time.ParseDuration(configRun.ReportInterval)
	if err != nil {
		log.Println(err)
	}
	tickerPoll := time.NewTicker(pollInterval)
	tickerReport := time.NewTicker(reportInterval)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	go func() {
		for {
			sig := <-sigs
			handleSignal(sig)
		}
	}()
	MetricsCurrent.InitMetrics()
	//todo добавить WG
	go func() {
		for range tickerPoll.C {
			go func() {
				MetricsCurrent.MuGauge.Lock()
				metricsupdate.UpdateGaugeMetricsRuntime(&MetricsCurrent)
				MetricsCurrent.MuGauge.Unlock()
				MetricsCurrent.MuCounter.Lock()
				MetricsCurrent.CounterMetric = metricsupdate.UpdateCounterMetrics("add", MetricsCurrent.CounterMetric)
				MetricsCurrent.MuCounter.Unlock()
			}()
			go func() {
				MetricsCurrent.MuGauge.Lock()
				metricsupdate.UpdateGaugeMetricsCPU(&MetricsCurrent)
				MetricsCurrent.MuGauge.Unlock()
			}()
		}
	}()
	go func() {
		for range tickerReport.C {
			metricssend.SendMetricJSON(&MetricsCurrent, configRun.Proto+configRun.Address, &configRun)
			metricssend.SendMetricsBatch(&MetricsCurrent, configRun.Proto+configRun.Address)
			MetricsCurrent.MuCounter.Lock()
			MetricsCurrent.CounterMetric = metricsupdate.UpdateCounterMetrics("init", MetricsCurrent.CounterMetric)
			MetricsCurrent.MuCounter.Unlock()
		}
	}()
	select {}
}
