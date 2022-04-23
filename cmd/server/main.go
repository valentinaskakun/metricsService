package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

//вынести структуру в модуль (?)
type Metrics struct {
	gaugeMetric   map[string]float64
	counterMetric map[string]int64
	timeMetric    time.Time
	sync.RWMutex
}

var serverToGetAddress = "127.0.0.1:8080"

var MetricsRun Metrics

func listMetricsAll(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "METRICS GAUGE:")
	for key, value := range MetricsRun.gaugeMetric {
		fmt.Fprintln(w, key, value)
	}
	fmt.Fprintln(w, "METRICS COUNTER:")
	for key, value := range MetricsRun.counterMetric {
		fmt.Fprintln(w, key, value)
	}
}
func listMetric(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")
	if metricType == "gauge" {
		if val, ok := MetricsRun.gaugeMetric[metricName]; ok {
			fmt.Fprintln(w, val)
		} else {
			w.WriteHeader(404)
		}
	} else if metricType == "counter" {
		if val, ok := MetricsRun.counterMetric[metricName]; ok {
			fmt.Fprintln(w, val)
		} else {
			w.WriteHeader(404)
		}
	} else {
		w.WriteHeader(501)
	}
}

//разнести по типам (?)
func updateMetrics(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")
	metricValue := chi.URLParam(r, "metricValue")
	if metricType == "gauge" {
		valParsed, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			w.WriteHeader(400)
		} else {
			MetricsRun.Lock()
			MetricsRun.gaugeMetric[metricName] = valParsed
			MetricsRun.Unlock()
		}
	} else if metricType == "counter" {
		valParsed, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			w.WriteHeader(400)
		} else {
			MetricsRun.Lock()
			MetricsRun.counterMetric[metricName] += valParsed
			MetricsRun.Unlock()
		}
	} else {
		w.WriteHeader(501)
	}
}

func main() {
	MetricsRun.gaugeMetric = make(map[string]float64)
	MetricsRun.counterMetric = make(map[string]int64)
	r := chi.NewRouter()
	r.Get("/", listMetricsAll)
	r.Route("/update", func(r chi.Router) {
		r.Route("/{metricType}", func(r chi.Router) {
			r.Post("/{metricName}/{metricValue}", updateMetrics)
		})
	})
	r.Route("/value", func(r chi.Router) {
		r.Route("/{metricType}", func(r chi.Router) {
			r.Get("/{metricName}", listMetric)
		})
	})
	log.Fatal(http.ListenAndServe(serverToGetAddress, r))
}
