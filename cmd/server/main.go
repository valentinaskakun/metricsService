package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"strconv"
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
	//fmt.Fprintln(w, r.RequestURI)
	//w.WriteHeader(http.StatusOK)
	metricType := chi.URLParam(r, "metricType")
	//fmt.Fprintln(w, "bogdan update:"+chi.URLParam(r, "metricType"))
	if metricType == "gauge" {
		fmt.Fprintln(w, "bogdan gauge:"+chi.URLParam(r, "metricType"))
		MetricsRun.gaugeMetric[chi.URLParam(r, "metricName")], _ = strconv.ParseFloat(chi.URLParam(r, "metricValue"), 64)
	} else if metricType == "counter" {
		fmt.Fprintln(w, "bogdan counter:"+chi.URLParam(r, "metricType"))
		MetricsRun.counterMetric[chi.URLParam(r, "metricName")], _ = strconv.ParseInt(chi.URLParam(r, "metricValue"), 10, 64)
	} else {
		fmt.Println("bogdan else:" + chi.URLParam(r, "metricType"))
		w.WriteHeader(http.StatusNotFound)
	}
	//splitURL := strings.Split(r.RequestURI, "/")
	//if len(splitURL) == 5 {
	//	if splitURL[2] == "gauge" {
	//		metricName := splitURL[3]
	//		metricValue, _ := strconv.ParseFloat(splitURL[4], 64)
	//		MetricsRun.gaugeMetric[metricName] = metricValue
	//	}
	//	if splitURL[2] == "counter" {
	//		metricName := splitURL[3]
	//		metricValue, _ := strconv.ParseInt(splitURL[4], 10, 64)
	//		MetricsRun.counterMetric[metricName] += metricValue
	//	}
	//}
}

func main() {
	MetricsRun.gaugeMetric = make(map[string]float64)
	MetricsRun.counterMetric = make(map[string]int64)
	r := chi.NewRouter()
	r.Get("/", listMetrics)
	r.Get("/update/{metricType}/{metricName}/{metricValue}", updateMetrics)
	log.Fatal(http.ListenAndServe(serverToGetAddress, r))
}
