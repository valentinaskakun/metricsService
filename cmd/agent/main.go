package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"runtime"
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
	//metricsCurrent []MetricGauge
	gaugeMetric   map[string]float64
	counterMetric map[string]int64
	//pollCount     int
	//randomValue   float64
	timeMetric time.Time
}

var pollInterval time.Duration = 2000    //Milliseconds
var reportInterval time.Duration = 10000 //Milliseconds
var serverToSendAddress = "127.0.0.1:8080"
var serverToSendProto = "http"

func getCurrentRuntimeMemStats() (currentMemStats runtime.MemStats) {
	runtime.ReadMemStats(&currentMemStats)
	return currentMemStats
}
func updateValuesRuntime(currentMemStats runtime.MemStats, metricsForUpdate *Metrics) {
	metricsForUpdate.gaugeMetric["alloc"] = float64(currentMemStats.Alloc)
	metricsForUpdate.gaugeMetric["buckHashSys"] = float64(currentMemStats.BuckHashSys)
	return
}
func updateMetrics(metricsToUpdate *Metrics) {
	updateValuesRuntime(getCurrentRuntimeMemStats(), metricsToUpdate)
	metricsToUpdate.counterMetric["pollCount"] += 1
	metricsToUpdate.gaugeMetric["randomValue"] = rand.Float64()
	metricsToUpdate.timeMetric = time.Now()
	fmt.Println(time.Now(), "updating")
	return
}
func sendMetrics(metricsToSend *Metrics) {
	for key, value := range metricsToSend.gaugeMetric {
		fmt.Println("sendMetrics gauge", key, value)
		//sendPOST("update", "gauge", key, fmt.Sprintf("%f", value))
	}
	for key, value := range metricsToSend.counterMetric {
		fmt.Println("sendMetrics counter", key, value)
		//sendPOST("update", "counter", key, strconv.FormatInt(value, 10))
	}
	fmt.Println("counter", metricsToSend.counterMetric["pollCount"])
	return
}
func sendPOST(urlAction string, urlMetricType string, urlMetricKey string, urlMetricValue string) {
	//http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
	url := serverToSendProto + "://" + serverToSendAddress + "/" + urlAction + "/"
	url += urlMetricType + "/" + urlMetricKey + "/" + urlMetricValue
	method := "POST"
	//contentType := "Content-Type: text/plain"
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	req.Header.Add("Content-Type", "Content-Type: text/plain")
	if err != nil {
		fmt.Println(err)
		return
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
	return
}
func inBackgroundMetrics(tickerToBackground *time.Ticker, metricsToBackground *Metrics, functionToBackground func(*Metrics)) {
	for _ = range tickerToBackground.C {
		functionToBackground(metricsToBackground)
		fmt.Println(tickerToBackground, " ticking")
	}
}
func handleSignal(signal os.Signal) {
	fmt.Println("* Got:", signal)
	os.Exit(-1)
}
func main() {
	//currentMemStats := getCurrentRuntimeMemStats()

	tickerPoll := time.NewTicker(time.Millisecond * pollInterval)
	tickerReport := time.NewTicker(time.Millisecond * reportInterval)
	metricsAgent := new(Metrics)
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
	for {
	}
}
