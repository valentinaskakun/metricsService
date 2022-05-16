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
type GaugeMemory struct {
	metric map[string]float64
	mutex  sync.Mutex
}

type CounterMemory struct {
	metric map[string]int64
	mutex  sync.Mutex
}

var (
	GaugeMetric   GaugeMemory
	CounterMetric CounterMemory
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

var serverToGetAddress = "127.0.0.1:8080"

var MetricsRun Metrics

//todo: вынести хэндлеры в интернал
func listMetricsAll(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "METRICS GAUGE:")
	//todo: нужно ли добавлять RLock
	for key, value := range GaugeMetric.metric {
		fmt.Fprintln(w, key, value)
	}
	fmt.Fprintln(w, "METRICS COUNTER:")
	for key, value := range CounterMetric.metric {
		fmt.Fprintln(w, key, value)
	}
}
func listMetric(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")
	if metricType == "gauge" {
		if val, ok := GaugeMetric.metric[metricName]; ok {
			fmt.Fprintln(w, val)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	} else if metricType == "counter" {
		if val, ok := CounterMetric.metric[metricName]; ok {
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
		if _, ok := GaugeMetric.metric[metricReq.ID]; ok {
			metricRes.ID, metricRes.MType, metricRes.Delta = metricReq.ID, metricReq.MType, metricReq.Delta
			valueRes := GaugeMetric.metric[metricReq.ID]
			metricRes.Value = &valueRes
		} else {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	} else if metricReq.MType == "counter" {
		if _, ok := CounterMetric.metric[metricReq.ID]; ok {
			metricRes.ID, metricRes.MType, metricRes.Value = metricReq.ID, metricReq.MType, metricReq.Value
			valueRes := CounterMetric.metric[metricReq.ID]
			metricRes.Delta = &valueRes
		} else {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
		return
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
			GaugeMetric.metric[metricName] = valParsed
			MetricsRun.muGauge.Unlock()
		}
	} else if metricType == "counter" {
		valParsed, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			MetricsRun.muCounter.Lock()
			CounterMetric.metric[metricName] += valParsed
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
		//MetricsRun.muGauge.Lock()
		GaugeMetric.metric[metricReq.ID] = *metricReq.Value
		//MetricsRun.muGauge.Unlock()
	} else if metricReq.MType == "counter" {
		//MetricsRun.muCounter.Lock()
		CounterMetric.metric[metricReq.ID] += *metricReq.Delta
		//MetricsRun.muCounter.Unlock()
	} else {
		w.WriteHeader(http.StatusNotImplemented)
	}
	w.WriteHeader(http.StatusOK)
	resBody, _ := json.Marshal("{}")
	w.Write(resBody)

	//if response, err := json.Marshal(); err != nil {
	//	http.Error(w, err.Error(), http.StatusUnprocessableEntity)
	//	return
	//} else {
	//	w.WriteHeader(http.StatusOK)
	//	w.Write(response)
	//}
}
func main() {
	GaugeMetric.metric = make(map[string]float64)
	CounterMetric.metric = make(map[string]int64)
	r := chi.NewRouter()
	r.Get("/", listMetricsAll)
	r.Route("/update", func(r chi.Router) {
		r.Post("/", updateMetricJSON)
		r.Post("/{metricType}/{metricName}/{metricValue}", updateMetrics)
	})
	r.Route("/value", func(r chi.Router) {
		r.Post("/", listMetricJSON)
		r.Get("/{metricType}/{metricName}", listMetric)
	})
	log.Fatal(http.ListenAndServe(serverToGetAddress, r))
}
