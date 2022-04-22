package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"time"
)

type MetricGauge struct {
	metricName  string
	metricValue float64
}
type MetricCount struct {
	metricName  string
	metricValue int
}
type Metrics struct {
	gaugeMetric   map[string]float64
	counterMetric map[string]int64
	timeMetric    time.Time
}

var pollInterval time.Duration = 2000    //Milliseconds
var reportInterval time.Duration = 10000 //Milliseconds
var serverToSendAddress = "127.0.0.1:8080"
var serverToSendProto = "http"

//добавить возможность вытаскивать конфиг только по нужным метрикам (?)
//var configMetrics = []string{"alloc", "buckHashSys"}

func getCurrentRuntimeMemStats() (currentMemStats runtime.MemStats) {
	runtime.ReadMemStats(&currentMemStats)
	return currentMemStats
}
func updateValuesRuntime(currentMemStats runtime.MemStats, metricsForUpdate Metrics) {
	metricsForUpdate.gaugeMetric["Alloc"] = float64(currentMemStats.Alloc)
	metricsForUpdate.gaugeMetric["BuckHashSys"] = float64(currentMemStats.BuckHashSys)
	metricsForUpdate.gaugeMetric["Frees"] = float64(currentMemStats.Frees)
	metricsForUpdate.gaugeMetric["GCCPUFraction"] = float64(currentMemStats.GCCPUFraction)
	metricsForUpdate.gaugeMetric["GCSys"] = float64(currentMemStats.GCSys)
	metricsForUpdate.gaugeMetric["HeapAlloc"] = float64(currentMemStats.HeapAlloc)
	metricsForUpdate.gaugeMetric["HeapIdle"] = float64(currentMemStats.HeapIdle)
	metricsForUpdate.gaugeMetric["HeapInuse"] = float64(currentMemStats.HeapInuse)
	metricsForUpdate.gaugeMetric["HeapObjects"] = float64(currentMemStats.HeapObjects)
	metricsForUpdate.gaugeMetric["HeapReleased"] = float64(currentMemStats.HeapReleased)
	metricsForUpdate.gaugeMetric["HeapSys"] = float64(currentMemStats.HeapSys)
	metricsForUpdate.gaugeMetric["LastGC"] = float64(currentMemStats.LastGC)
	metricsForUpdate.gaugeMetric["Lookups"] = float64(currentMemStats.Lookups)
	metricsForUpdate.gaugeMetric["MCacheInuse"] = float64(currentMemStats.MCacheInuse)
	metricsForUpdate.gaugeMetric["MCacheSys"] = float64(currentMemStats.MCacheSys)
	metricsForUpdate.gaugeMetric["MSpanInuse"] = float64(currentMemStats.MSpanInuse)
	metricsForUpdate.gaugeMetric["MSpanSys"] = float64(currentMemStats.MSpanSys)
	metricsForUpdate.gaugeMetric["Mallocs"] = float64(currentMemStats.Mallocs)
	metricsForUpdate.gaugeMetric["NextGC"] = float64(currentMemStats.NextGC)
	metricsForUpdate.gaugeMetric["NumForcedGC"] = float64(currentMemStats.NumForcedGC)
	metricsForUpdate.gaugeMetric["NumGC"] = float64(currentMemStats.NumGC)
	metricsForUpdate.gaugeMetric["OtherSys"] = float64(currentMemStats.OtherSys)
	metricsForUpdate.gaugeMetric["PauseTotalNs"] = float64(currentMemStats.PauseTotalNs)
	metricsForUpdate.gaugeMetric["StackInuse"] = float64(currentMemStats.StackInuse)
	metricsForUpdate.gaugeMetric["StackSys"] = float64(currentMemStats.StackSys)
	metricsForUpdate.gaugeMetric["Sys"] = float64(currentMemStats.Sys)
	metricsForUpdate.gaugeMetric["TotalAlloc"] = float64(currentMemStats.TotalAlloc)
}
func updateMetrics(metricsToUpdate Metrics) {
	updateValuesRuntime(getCurrentRuntimeMemStats(), metricsToUpdate)
	metricsToUpdate.counterMetric["pollCount"] += 1
	metricsToUpdate.gaugeMetric["randomValue"] = rand.Float64()
	metricsToUpdate.timeMetric = time.Now()
	fmt.Println(time.Now(), "updating")
}
func sendMetrics(metricsToSend Metrics) {
	for key, value := range metricsToSend.gaugeMetric {
		//fmt.Println("sendMetrics gauge", key, value)
		sendPOST("update", "gauge", key, fmt.Sprintf("%f", value))
		//fmt.Println(test.StatusCode)
	}
	for key, value := range metricsToSend.counterMetric {
		//fmt.Println("sendMetrics counter", key, value)
		sendPOST("update", "counter", key, strconv.FormatInt(value, 10))
	}
	//fmt.Println("counter", metricsToSend.counterMetric["pollCount"])
}
func sendPOST(urlAction string, urlMetricType string, urlMetricKey string, urlMetricValue string) {
	//http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
	url := serverToSendProto + "://" + serverToSendAddress + "/" + urlAction + "/"
	url += urlMetricType + "/" + urlMetricKey + "/" + urlMetricValue
	method := "POST"
	//contentType := "Content-Type: text/plain"
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Content-Type", "Content-Type: text/plain")
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()
}
func inBackgroundMetrics(tickerToBackground *time.Ticker, metricsToBackground *Metrics, functionToBackground func(*Metrics)) {
	for range tickerToBackground.C {
		functionToBackground(metricsToBackground)
		fmt.Println(tickerToBackground, " ticking")
	}
}
func handleSignal(signal os.Signal) {
	fmt.Println("* Got:", signal)
	os.Exit(-1)
}
func main() {
	tickerPoll := time.NewTicker(time.Millisecond * pollInterval)
	tickerReport := time.NewTicker(time.Millisecond * reportInterval)
	var metricsAgent Metrics
	metricsAgent.gaugeMetric = make(map[string]float64)
	metricsAgent.counterMetric = make(map[string]int64)
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
				updateMetrics(metricsAgent)
			case <-tickerReport.C:
				sendMetrics(metricsAgent)
			}
		}
	}()
	select {}
}
