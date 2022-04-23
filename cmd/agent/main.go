package main

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/valentinaskakun/metricsService.git/internal/metricsruntime"
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
	gaugeMetric   map[string]float64
	counterMetric map[string]int64
	timeMetric    time.Time
	sync.RWMutex
}

var pollInterval time.Duration = 2000    //Milliseconds
var reportInterval time.Duration = 10000 //Milliseconds
var serverToSendProto = "http://"
var serverToSend = serverToSendProto + "127.0.0.1:8080"
var metricsListConfig = map[string]bool{"Alloc": true, "BuckHashSys": true, "Frees": true, "GCCPUFraction": true, "GCSys": true, "HeapAlloc": true, "HeapIdle": true, "HeapInuse": true, "HeapObjects": true, "HeapReleased": true, "HeapSys": true, "LastGC": true, "Lookups": true, "MCacheInuse": true, "MCacheSys": true, "MSpanInuse": true, "MSpanSys": true, "Mallocs": true, "NextGC": true, "NumForcedGC": true, "NumGC": true, "OtherSys": true, "PauseTotalNs": true, "StackInuse": true, "StackSys": true, "Sys": true, "TotalAlloc": true, "pollCount": true}
var MetricsCurrent Metrics

//добавить обработку ошибок
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
	metricsCounterUpdated["pollCount"] = 1
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
	fmt.Println(action)
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
	go func() {
		for {
			select {
			case <-tickerPoll.C:
				MetricsCurrent.Lock()
				MetricsCurrent.gaugeMetric = updateGaugeMetrics()
				MetricsCurrent.counterMetric = updateCounterMetrics("add", MetricsCurrent.counterMetric)
				MetricsCurrent.Unlock()
			}
		}
	}()
	go func() {
		for {
			select {
			case <-tickerReport.C:
				sendMetrics(&MetricsCurrent, serverToSend)
				MetricsCurrent.Lock()
				MetricsCurrent.counterMetric = updateCounterMetrics("init", MetricsCurrent.counterMetric)
				MetricsCurrent.Unlock()
			}
		}
	}()
	select {}
}
