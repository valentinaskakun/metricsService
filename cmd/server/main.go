package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Metrics struct {
	gaugeMetric   map[string]float64
	counterMetric map[string]int64
	timeMetric    time.Time
}

var serverToGetAddress = "127.0.0.1:8080"

var MetricsRun Metrics

//var serverToGetProto = "http"
func listMetrics(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "METRICS:")
	for key, value := range MetricsRun.gaugeMetric {
		fmt.Fprintln(w, key, value)
	}
	for key, value := range MetricsRun.counterMetric {
		fmt.Fprintln(w, key, value)
	}

}
func updateMetrics(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(r.RequestURI + "\n"))
	var splitUrl []string
	splitUrl = strings.Split(r.RequestURI, "/")
	if len(splitUrl) == 5 {
		if splitUrl[2] == "gauge" {
			metricName := splitUrl[3]
			metricValue, _ := strconv.ParseFloat(splitUrl[4], 64)
			MetricsRun.gaugeMetric[metricName] = metricValue
		}
		if splitUrl[2] == "counter" {
			metricName := splitUrl[3]
			metricValue, _ := strconv.ParseInt(splitUrl[4], 10, 64)
			MetricsRun.counterMetric[metricName] += metricValue
		}
	}
}

func main() {

	MetricsRun.gaugeMetric = make(map[string]float64)
	MetricsRun.counterMetric = make(map[string]int64)

	// маршрутизация запросов обработчику
	http.HandleFunc("/", listMetrics)
	http.HandleFunc("/update/", updateMetrics)
	// конструируем свой сервер
	server := &http.Server{
		Addr: serverToGetAddress,
	}

	server.ListenAndServe()
}
