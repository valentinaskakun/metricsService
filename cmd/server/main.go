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
	//fmt.Fprintln(w, "METRIC:")
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")
	//fmt.Fprintln(w, "bogdan update:"+chi.URLParam(r, "metricType"))
	if metricType == "gauge" {
		fmt.Println("bogdan listMetric gauge:" + chi.URLParam(r, "metricType"))
		if val, ok := MetricsRun.gaugeMetric[metricName]; ok {
			fmt.Fprintln(w, val)
		} else {
			w.WriteHeader(404)
		}
	} else if metricType == "counter" {
		fmt.Println("bogdan listMetric counter:" + chi.URLParam(r, "metricType"))
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
	fmt.Println("ya zdes", r.RequestURI)
	//w.WriteHeader(http.StatusOK)
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")
	metricValue := chi.URLParam(r, "metricValue")
	//fmt.Fprintln(w, "bogdan update:"+chi.URLParam(r, "metricType"))
	if metricType == "gauge" {
		fmt.Println("bogdan gauge:" + chi.URLParam(r, "metricType"))
		valParsed, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			w.WriteHeader(400)
		} else {
			MetricsRun.gaugeMetric[metricName] = valParsed
		}
	} else if metricType == "counter" {
		fmt.Println("bogdan counter:" + chi.URLParam(r, "metricType"))
		valParsed, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			w.WriteHeader(400)
		} else {
			MetricsRun.counterMetric[metricName] = valParsed
		}
	} else {
		fmt.Println("bogdan else:" + chi.URLParam(r, "metricType"))
		w.WriteHeader(501)
	}
}

func main() {
	MetricsRun.gaugeMetric = make(map[string]float64)
	MetricsRun.counterMetric = make(map[string]int64)
	r := chi.NewRouter()
	r.Post("/", listMetricsAll)
	r.Route("/update", func(r chi.Router) {
		r.Route("/{metricType}", func(r chi.Router) {
			r.Post("/{metricName}/{metricValue}", updateMetrics)
		})
	})
	r.Route("/{metricType}", func(r chi.Router) {
		r.Get("/{metricName}/", listMetric)
	})
	//r.Get("/update/{metricType}/{metricName}/{metricValue}", updateMetrics)
	//r.Route("/{carID}", func(r chi.Router) {
	log.Fatal(http.ListenAndServe(serverToGetAddress, r))
}
