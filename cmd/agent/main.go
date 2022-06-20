package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/valentinaskakun/metricsService/internal/config"
	"github.com/valentinaskakun/metricsService/internal/metricsruntime"
	"github.com/valentinaskakun/metricsService/internal/storage"
	"math/rand"
	"net/url"
	"os"
	"os/signal"
	"path"
	"strconv"
	"syscall"
	"time"
)

//todo: навести порядок
const (
	pollIntervalConst   = 2000
	reportIntervalConst = 4000
)

//var pollInterval time.Duration = pollIntervalConst     //Milliseconds
//var reportInterval time.Duration = reportIntervalConst //Milliseconds
var metricsListConfig = map[string]bool{"Alloc": true, "BuckHashSys": true, "Frees": true, "GCCPUFraction": true, "GCSys": true, "HeapAlloc": true, "HeapIdle": true, "HeapInuse": true, "HeapObjects": true, "HeapReleased": true, "HeapSys": true, "LastGC": true, "Lookups": true, "MCacheInuse": true, "MCacheSys": true, "MSpanInuse": true, "MSpanSys": true, "Mallocs": true, "NextGC": true, "NumForcedGC": true, "NumGC": true, "OtherSys": true, "PauseTotalNs": true, "StackInuse": true, "StackSys": true, "Sys": true, "TotalAlloc": true, "PollCount": true}
var MetricsCurrent storage.Metrics
var serverToSendProto = "http://"

//todo: добавить обработку ошибок
//todo: закинуть все в модуль datamanipulation
func updateGaugeMetrics() (metricsGaugeUpdated map[string]float64) {
	metricsGaugeUpdated = make(map[string]float64)
	tempCurrentMemStatsMetrics := metricsruntime.GetCurrentValuesRuntimeGauge()
	for key, value := range tempCurrentMemStatsMetrics {
		if _, ok := metricsListConfig[key]; ok {
			metricsGaugeUpdated[key] = value
		}
	}
	metricsGaugeUpdated["RandomValue"] = rand.Float64()
	return metricsGaugeUpdated
}
func updateCounterMetrics(action string, metricsCounterToUpdate map[string]int64) (metricsCounterUpdated map[string]int64) {
	metricsCounterUpdated = make(map[string]int64)
	if _, ok := metricsCounterToUpdate["PollCount"]; !ok {
		metricsCounterToUpdate = make(map[string]int64)
		metricsCounterToUpdate["PollCount"] = 0
	}
	switch {
	case action == "add":
		for key, value := range metricsCounterToUpdate {
			if _, ok := metricsListConfig[key]; ok {
				metricsCounterUpdated[key] = value + 1
			}
		}
	case action == "init":
		for key := range metricsCounterToUpdate {
			if _, ok := metricsListConfig[key]; ok {
				metricsCounterUpdated[key] = 0
			}
		}
	}
	return metricsCounterUpdated
}

func sendMetrics(metricsToSend *storage.Metrics, serverToSendLink string) {
	if metricsToSend.CounterMetric["PollCount"] != 0 {
		for key, value := range metricsToSend.GaugeMetric {
			urlStr, err := url.Parse(serverToSendLink)
			if err != nil {
				fmt.Println(err)
				return
			}
			urlStr.Path = path.Join(urlStr.Path, "update", "gauge", key, fmt.Sprintf("%f", value))
			sendPOST(urlStr.String())
		}
		for key, value := range metricsToSend.CounterMetric {
			urlStr, err := url.Parse(serverToSendLink)
			if err != nil {
				fmt.Println(err)
				return
			}
			urlStr.Path = path.Join(urlStr.Path, "update", "counter", key, strconv.FormatInt(value, 10))
			sendPOST(urlStr.String())
		}
	}
}
func sendPOST(urlString string) {
	client := resty.New()
	_, err := client.R().
		SetHeader("Content-Type", "Content-Type: text/plain").
		Post(urlString)
	if err != nil {
		fmt.Println("ERROR POST " + urlString)
		return
	}
}

//todo: переделать использование url server path
func sendMetricJSON(metricsToSend *storage.Metrics, serverToSendLink string) {
	if metricsToSend.CounterMetric["PollCount"] != 0 {
		urlStr, _ := url.Parse(serverToSendLink)
		urlStr.Path = path.Join(urlStr.Path, "update")
		client := resty.New()
		client.R().
			SetHeader("Content-Type", "Content-Type: application/json")
		for key, value := range metricsToSend.GaugeMetric {
			metricToSend, err := json.Marshal(storage.MetricsJSON{ID: key, MType: "gauge", Value: &value})
			if err != nil {
				fmt.Println(err)
				return
			}
			client.R().
				SetBody(metricToSend).
				Post(urlStr.String())
		}
		for key, value := range metricsToSend.CounterMetric {
			metricToSend, err := json.Marshal(storage.MetricsJSON{ID: key, MType: "counter", Delta: &value})
			if err != nil {
				fmt.Println(err)
				return
			}
			client.R().
				SetBody(metricToSend).
				Post(urlStr.String())
		}
	} else {
		fmt.Println("ERROR: Something went wrong while sendingMetricJSON")
	}
}
func handleSignal(signal os.Signal) {
	fmt.Println("* Got:", signal)
	os.Exit(-1)
}
func main() {
	configRun, _ := config.LoadConfigAgent()
	fmt.Println(configRun)
	pollInterval, _ := time.ParseDuration(configRun.PollInterval)
	reportInterval, _ := time.ParseDuration(configRun.ReportInterval)
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
	//todo добавить WG
	go func() {
		for range tickerPoll.C {
			//todo проверить локи, они рааботают вообще норм или нет
			MetricsCurrent.MuGauge.Lock()
			MetricsCurrent.GaugeMetric = updateGaugeMetrics()
			MetricsCurrent.MuGauge.Unlock()
			MetricsCurrent.MuCounter.Lock()
			MetricsCurrent.CounterMetric = updateCounterMetrics("add", MetricsCurrent.CounterMetric)
			MetricsCurrent.MuCounter.Unlock()
		}
	}()
	go func() {
		for range tickerReport.C {
			sendMetricJSON(&MetricsCurrent, serverToSendProto+configRun.Address)
			MetricsCurrent.MuCounter.Lock()
			MetricsCurrent.CounterMetric = updateCounterMetrics("init", MetricsCurrent.CounterMetric)
			MetricsCurrent.MuCounter.Unlock()
		}
	}()
	select {}
}
