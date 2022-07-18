package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/rs/zerolog"

	"github.com/valentinaskakun/metricsService/internal/config"
	"github.com/valentinaskakun/metricsService/internal/metricscollect"
	"github.com/valentinaskakun/metricsService/internal/storage"

	"github.com/go-resty/resty/v2"
)

//todo: навести порядок
const (
	pollIntervalConst   = 2000
	reportIntervalConst = 4000
)

//var pollInterval time.Duration = pollIntervalConst     //Milliseconds
//var reportInterval time.Duration = reportIntervalConst //Milliseconds
var metricsListConfig = map[string]bool{"TotalMemory": true, "FreeMemory": true, "CPUutilization1": true, "Alloc": true, "BuckHashSys": true, "Frees": true, "GCCPUFraction": true, "GCSys": true, "HeapAlloc": true, "HeapIdle": true, "HeapInuse": true, "HeapObjects": true, "HeapReleased": true, "HeapSys": true, "LastGC": true, "Lookups": true, "MCacheInuse": true, "MCacheSys": true, "MSpanInuse": true, "MSpanSys": true, "Mallocs": true, "NextGC": true, "NumForcedGC": true, "NumGC": true, "OtherSys": true, "PauseTotalNs": true, "StackInuse": true, "StackSys": true, "Sys": true, "TotalAlloc": true, "PollCount": true}
var MetricsCurrent storage.Metrics
var serverToSendProto = "http://"

//todo: добавить обработку ошибок
//todo: закинуть все в модуль datamanipulation
func updateGaugeMetricsRuntime() (metricsGaugeUpdated map[string]float64) {
	metricsGaugeUpdated = make(map[string]float64)
	tempCurrentMemStatsMetrics := metricscollect.GetCurrentValuesRuntimeGauge()
	for key, value := range tempCurrentMemStatsMetrics {
		if _, ok := metricsListConfig[key]; ok {
			metricsGaugeUpdated[key] = value
		}
	}
	metricsGaugeUpdated["RandomValue"] = rand.Float64()
	return metricsGaugeUpdated
}
func updateGaugeMetricsCPU() (metricsGaugeUpdated map[string]float64) {
	metricsGaugeUpdated = make(map[string]float64)
	tempCurrentCPUStatsMetrics := metricscollect.GetCurrentValuesGOpsGauge()
	for key, value := range tempCurrentCPUStatsMetrics {
		metricsGaugeUpdated[key] = value
	}
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

//todo: переделать использование url server path
//todo: добавить использование configRun-полей, вот этого монстра перенести из мэйна
func sendMetricJSON(metricsToSend *storage.Metrics, serverToSendLink string, configRun *config.ConfAgent) {
	log := zerolog.New(os.Stdout)
	if metricsToSend.CounterMetric["PollCount"] != 0 {
		urlStr, _ := url.Parse(serverToSendLink)
		urlStr.Path = path.Join(urlStr.Path, "update")
		client := resty.New()
		client.R().
			SetHeader("Content-Type", "Content-Type: application/json")
		for key, value := range metricsToSend.GaugeMetric {
			metricToSend, err := json.Marshal(storage.MetricsJSON{ID: key, MType: "gauge", Value: &value})
			if err != nil {
				log.Warn().Msg(err.Error())
				return
			}
			if len(configRun.Key) > 0 {
				//todo: переделать функцию хэш с нормальными аргументами
				hashValue := config.Hash(fmt.Sprintf("%s:gauge:%f", key, value), configRun.Key)
				metricToSend, err = json.Marshal(storage.MetricsJSON{ID: key, MType: "gauge", Value: &value, Hash: hashValue})
				if err != nil {
					log.Warn().Msg(err.Error())
					return
				}
			}
			_, err = client.R().
				SetBody(metricToSend).
				Post(urlStr.String())
			if err != nil {
				log.Warn().Msg(err.Error())
				return
			}
		}
		for key, value := range metricsToSend.CounterMetric {
			metricToSend, err := json.Marshal(storage.MetricsJSON{ID: key, MType: "counter", Delta: &value})
			if err != nil {
				log.Warn().Msg(err.Error())
				return
			}
			if len(configRun.Key) > 0 {
				hashValue := config.Hash(fmt.Sprintf("%s:counter:%d", key, value), configRun.Key)
				metricToSend, err = json.Marshal(storage.MetricsJSON{ID: key, MType: "counter", Delta: &value, Hash: hashValue})
				if err != nil {
					log.Warn().Msg(err.Error())
					return
				}
			}
			_, err = client.R().
				SetBody(metricToSend).
				Post(urlStr.String())
			if err != nil {
				log.Warn().Msg(err.Error())
				return
			}
		}
	} else {
		fmt.Println("ERROR: Something went wrong while sendingMetricJSON")
	}
}
func sendMetricsBatch(metricsToSend *storage.Metrics, serverToSendLink string) {
	log := zerolog.New(os.Stdout)
	var metricsBatch []storage.MetricsJSON
	if metricsToSend.CounterMetric["PollCount"] != 0 {
		urlStr, _ := url.Parse(serverToSendLink)
		urlStr.Path = path.Join(urlStr.Path, "updates")
		client := resty.New()
		client.R().
			SetHeader("Content-Type", "Content-Type: application/json")
		for key, value := range metricsToSend.GaugeMetric {
			newVal := value
			metricToSend := storage.MetricsJSON{ID: key, MType: "gauge", Value: &newVal}
			metricsBatch = append(metricsBatch, metricToSend)
		}
		for key, value := range metricsToSend.CounterMetric {
			newVal := value
			metricToSend := storage.MetricsJSON{ID: key, MType: "counter", Delta: &newVal}
			metricsBatch = append(metricsBatch, metricToSend)
		}
		if len(metricsBatch) > 0 {
			metricsPrepared, err := json.Marshal(metricsBatch)
			if err != nil {
				log.Warn().Msg(err.Error())
				return
			}
			_, err = client.R().
				SetBody(metricsPrepared).
				Post(urlStr.String() + "/")
			if err != nil {
				log.Warn().Msg(err.Error())
				return
			}
		} else {
			return
		}

	} else {
		log.Info().Msg("ERROR: Something went wrong while sendingMetricJSON")
	}
}
func handleSignal(signal os.Signal) {
	fmt.Println("* Got:", signal)
	os.Exit(-1)
}
func main() {
	//test := metricscollect.GetCurrentValuesGOpsGauge()
	//fmt.Println(test)
	configRun, _ := config.LoadConfigAgent()
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
			go func() {
				MetricsCurrent.MuGauge.Lock()
				MetricsCurrent.GaugeMetric = updateGaugeMetricsRuntime()
				MetricsCurrent.MuGauge.Unlock()
				MetricsCurrent.MuCounter.Lock()
				//todo: переделать add по-человечески
				MetricsCurrent.CounterMetric = updateCounterMetrics("add", MetricsCurrent.CounterMetric)
				MetricsCurrent.MuCounter.Unlock()
			}()
			go func() {
				MetricsCurrent.MuGauge.Lock()
				MetricsCurrent.GaugeMetric = updateGaugeMetricsCPU()
				MetricsCurrent.MuGauge.Unlock()
			}()
		}
	}()
	go func() {
		for range tickerReport.C {
			//todo: оставить только config
			sendMetricJSON(&MetricsCurrent, serverToSendProto+configRun.Address, &configRun)
			sendMetricsBatch(&MetricsCurrent, serverToSendProto+configRun.Address)
			MetricsCurrent.MuCounter.Lock()
			MetricsCurrent.CounterMetric = updateCounterMetrics("init", MetricsCurrent.CounterMetric)
			MetricsCurrent.MuCounter.Unlock()
		}
	}()
	select {}
}
