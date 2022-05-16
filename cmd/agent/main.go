package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/valentinaskakun/metricsService/internal/metricsruntime"
	"math/rand"
	"net/url"
	"os"
	"os/signal"
	"path"
	"strconv"
	"sync"
	"syscall"
	"time"
)

type Metrics struct {
	muGauge       sync.RWMutex
	gaugeMetric   map[string]float64
	muCounter     sync.RWMutex
	counterMetric map[string]int64
	timeMetric    time.Time
}

type MetricsJSON struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

const (
	pollIntervalConst   = 2000
	reportIntervalConst = 4000
)

var pollInterval time.Duration = pollIntervalConst     //Milliseconds
var reportInterval time.Duration = reportIntervalConst //Milliseconds
var serverToSendProto = "http://"
var serverToSend = serverToSendProto + "127.0.0.1:8080"
var metricsListConfig = map[string]bool{"Alloc": true, "BuckHashSys": true, "Frees": true, "GCCPUFraction": true, "GCSys": true, "HeapAlloc": true, "HeapIdle": true, "HeapInuse": true, "HeapObjects": true, "HeapReleased": true, "HeapSys": true, "LastGC": true, "Lookups": true, "MCacheInuse": true, "MCacheSys": true, "MSpanInuse": true, "MSpanSys": true, "Mallocs": true, "NextGC": true, "NumForcedGC": true, "NumGC": true, "OtherSys": true, "PauseTotalNs": true, "StackInuse": true, "StackSys": true, "Sys": true, "TotalAlloc": true, "pollCount": true}
var MetricsCurrent Metrics

//todo: добавить обработку ошибок
func updateGaugeMetrics() (metricsGaugeUpdated map[string]float64) {
	metricsGaugeUpdated = make(map[string]float64)
	tempCurrentMemStatsMetrics := metricsruntime.GetCurrentValuesRuntimeGauge()
	for key, value := range tempCurrentMemStatsMetrics {
		if _, ok := metricsListConfig[key]; ok {
			metricsGaugeUpdated[key] = value
		}
	}
	metricsGaugeUpdated["randomValue"] = rand.Float64()
	return metricsGaugeUpdated
}
func updateCounterMetrics(action string, metricsCounterToUpdate map[string]int64) (metricsCounterUpdated map[string]int64) {
	metricsCounterUpdated = make(map[string]int64)
	if _, ok := metricsCounterToUpdate["pollCount"]; !ok {
		metricsCounterToUpdate = make(map[string]int64)
		metricsCounterToUpdate["pollCount"] = 0
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

func sendMetrics(metricsToSend *Metrics, serverToSendLink string) {
	if metricsToSend.counterMetric["pollCount"] != 0 {
		for key, value := range metricsToSend.gaugeMetric {
			urlStr, err := url.Parse(serverToSendLink)
			if err != nil {
				fmt.Println(err)
				return
			}
			urlStr.Path = path.Join(urlStr.Path, "update", "gauge", key, fmt.Sprintf("%f", value))
			sendPOST(urlStr.String())
		}
		for key, value := range metricsToSend.counterMetric {
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
func sendMetricJSON(metricsToSend *Metrics, serverToSendLink string) {
	if metricsToSend.counterMetric["pollCount"] != 0 {
		urlStr, _ := url.Parse(serverToSendLink)
		urlStr.Path = path.Join(urlStr.Path, "update")
		client := resty.New()
		client.R().
			SetHeader("Content-Type", "Content-Type: application/json")
		for key, value := range metricsToSend.gaugeMetric {
			metricToSend, err := json.Marshal(MetricsJSON{ID: key, MType: "gauge", Value: &value})
			if err != nil {
				fmt.Println(err)
				return
			}
			client.R().
				SetBody(metricToSend).
				Post(urlStr.String())
		}
		for key, value := range metricsToSend.counterMetric {
			metricToSend, err := json.Marshal(MetricsJSON{ID: key, MType: "counter", Delta: &value})
			if err != nil {
				fmt.Println(err)
				return
			}
			client.R().
				SetBody(metricToSend).
				Post(urlStr.String())
		}
	} else {
		fmt.Println("ERROR: " + "pollCounter is 0")
	}
}
func handleSignal(signal os.Signal) {
	fmt.Println("* Got:", signal)
	os.Exit(-1)
}
func main() {
	tickerPoll := time.NewTicker(time.Millisecond * pollInterval)
	tickerReport := time.NewTicker(time.Millisecond * reportInterval)
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
			MetricsCurrent.muGauge.Lock()
			MetricsCurrent.gaugeMetric = updateGaugeMetrics()
			MetricsCurrent.muGauge.Unlock()
			MetricsCurrent.muCounter.Lock()
			MetricsCurrent.counterMetric = updateCounterMetrics("add", MetricsCurrent.counterMetric)
			MetricsCurrent.muCounter.Unlock()
		}

	}()
	go func() {
		for range tickerReport.C {
			sendMetricJSON(&MetricsCurrent, serverToSend)
			MetricsCurrent.muCounter.Lock()
			MetricsCurrent.counterMetric = updateCounterMetrics("init", MetricsCurrent.counterMetric)
			MetricsCurrent.muCounter.Unlock()
		}
	}()
	select {}
}
