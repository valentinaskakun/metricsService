package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

//todo: вынести структуру в модуль + реализовать интерфейс хранения

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

var serverToGetAddress = "127.0.0.1:8080"

var MetricsRun Metrics

//todo: вынести хэндлеры в интернал
func listMetricsAll(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "METRICS GAUGE:")
	//todo: нужно ли добавлять RLock
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
			w.WriteHeader(http.StatusNotFound)
		}
	} else if metricType == "counter" {
		if val, ok := MetricsRun.counterMetric[metricName]; ok {
			fmt.Fprintln(w, val)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	} else {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func listMetricJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	metricReq := MetricsJSON{}
	metricRes := MetricsJSON{}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	if err := json.Unmarshal(body, &metricReq); err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	if metricReq.MType == "gauge" {
		if _, ok := MetricsRun.gaugeMetric[metricReq.ID]; ok {
			metricRes.ID, metricRes.MType = metricReq.ID, metricReq.MType
			valueRes := MetricsRun.gaugeMetric[metricReq.ID]
			metricRes.Value = &valueRes
		}
	} else if metricReq.MType == "counter" {
		if _, ok := MetricsRun.counterMetric[metricReq.ID]; ok {
			metricRes.ID, metricRes.MType = metricReq.ID, metricReq.MType
			valueRes := MetricsRun.counterMetric[metricReq.ID]
			metricRes.Delta = &valueRes
		}
	} else {
		w.WriteHeader(http.StatusNotImplemented)
	}
	if resBody, err := json.Marshal(metricRes); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write(resBody)
	}
}

func updateMetrics(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")
	metricValue := chi.URLParam(r, "metricValue")
	if metricType == "gauge" {
		valParsed, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			MetricsRun.muGauge.Lock()
			MetricsRun.gaugeMetric[metricName] = valParsed
			MetricsRun.muGauge.Unlock()
		}
	} else if metricType == "counter" {
		valParsed, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			MetricsRun.muCounter.Lock()
			MetricsRun.counterMetric[metricName] += valParsed
			MetricsRun.muCounter.Unlock()
		}
	} else {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func updateMetricJSON(w http.ResponseWriter, r *http.Request) {
	metricReq := MetricsJSON{}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	if err := json.Unmarshal(body, &metricReq); err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	if metricReq.MType == "gauge" {
		MetricsRun.muGauge.Lock()
		MetricsRun.gaugeMetric[metricReq.ID] = *metricReq.Value
		MetricsRun.muGauge.Unlock()
	} else if metricReq.MType == "counter" {
		MetricsRun.muCounter.Lock()
		MetricsRun.counterMetric[metricReq.ID] += *metricReq.Delta
		MetricsRun.muCounter.Unlock()
	} else {
		w.WriteHeader(http.StatusNotImplemented)
	}

	//if response, err := json.Marshal(); err != nil {
	//	http.Error(w, err.Error(), http.StatusUnprocessableEntity)
	//	return
	//} else {
	//	w.WriteHeader(http.StatusOK)
	//	w.Write(response)
	//}
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
		r.Post("/", updateMetricJSON)
	})
	r.Route("/value", func(r chi.Router) {
		r.Route("/{metricType}", func(r chi.Router) {
			r.Get("/{metricName}", listMetric)
		})
		r.Post("/", listMetricJSON)
	})
	log.Fatal(http.ListenAndServe(serverToGetAddress, r))
}
