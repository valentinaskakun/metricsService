package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
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
type MetricsGauge struct {
	ID    string
	Value float64
}
type MetricsCounter struct {
	ID    string
	Value int64
}

var (
	GaugeMetric   GaugeMemory
	CounterMetric CounterMemory
)

type Metrics1 struct {
	muGauge       sync.RWMutex
	gaugeMetric   map[string]float64
	muCounter     sync.RWMutex
	counterMetric map[string]int64
	timeMetric    time.Time
}

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

var serverToGetAddress = "127.0.0.1:8080"

//var MetricsRun Metrics

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
	if metricType == "counter" {
		if value, ok := CounterMetric.metric[metricName]; ok {
			fmt.Fprintln(w, value)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	} else if metricType == "gauge" {
		if value, ok := GaugeMetric.metric[metricName]; ok {
			fmt.Fprintln(w, value)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	} else {
		w.WriteHeader(http.StatusNotFound)

	}
}

func listMetricJSON(w http.ResponseWriter, r *http.Request) {
	var m Metrics
	err := json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if m.MType == "counter" {
		if value, ok := CounterMetric.metric[m.ID]; ok {
			m.Delta = &value
			render.JSON(w, r, m)
		} else {
			w.WriteHeader(http.StatusNotFound)

		}
	} else if m.MType == "gauge" {
		if value, ok := GaugeMetric.metric[m.ID]; ok {
			m.Value = &value
			render.JSON(w, r, m)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	} else {
		w.WriteHeader(http.StatusNotFound)

	}
}

func updateMetrics(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")
	metricValue := chi.URLParam(r, "metricValue")

	if metricType == "gauge" {
		var receivedMetric MetricsGauge
		var err error
		receivedMetric.ID = metricName
		receivedMetric.Value, err = strconv.ParseFloat(metricValue, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		GaugeMetric.mutex.Lock()
		GaugeMetric.metric[receivedMetric.ID] = receivedMetric.Value
		GaugeMetric.mutex.Unlock()

	} else if metricType == "counter" {
		var receivedMetric MetricsCounter
		receivedMetric.ID = metricName
		var err error
		receivedMetric.Value, err = strconv.ParseInt(metricValue, 0, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		previousValue := CounterMetric.metric[receivedMetric.ID]
		CounterMetric.mutex.Lock()
		CounterMetric.metric[receivedMetric.ID] = receivedMetric.Value + previousValue
		CounterMetric.mutex.Unlock()

	} else {
		w.WriteHeader(501)
	}
}

func updateMetricJSON(w http.ResponseWriter, r *http.Request) {
	metricReq := Metrics{}
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
