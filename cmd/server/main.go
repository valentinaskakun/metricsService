package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/spf13/viper"
	"net/http"
	"strconv"
	"sync"
)

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

func receiveMetric(w http.ResponseWriter, r *http.Request) {
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

func valueOfMetric(w http.ResponseWriter, r *http.Request) {
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
func listMetrics(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintln(w, "#########GAUGE METRICS#########")
	for key, value := range GaugeMetric.metric {
		fmt.Fprintln(w, key, value)

	}
	fmt.Fprintln(w, "#########COUNTER METRICS#########")
	for key, value := range CounterMetric.metric {
		fmt.Fprintln(w, key, value)

	}

}

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func receiveMetricJSON(w http.ResponseWriter, r *http.Request) {
	var m Metrics
	err := json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		//http.Error(w, err.Error(), http.StatusBadRequest)
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, m)
		return
	}
	if m.MType == "gauge" {
		if m.Value == nil {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			GaugeMetric.mutex.Lock()
			GaugeMetric.metric[m.ID] = *m.Value
			GaugeMetric.mutex.Unlock()
			render.JSON(w, r, m)
		}
	} else if m.MType == "counter" {
		if m.Delta == nil {
			w.WriteHeader(http.StatusBadRequest)
			CounterMetric.mutex.Unlock()
			render.JSON(w, r, m)
		} else {
			previousValue := CounterMetric.metric[m.ID]
			CounterMetric.mutex.Lock()
			CounterMetric.metric[m.ID] = *m.Delta + previousValue
			CounterMetric.mutex.Unlock()
			render.JSON(w, r, m)
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}

}

func valueOfMetricJSON(w http.ResponseWriter, r *http.Request) {
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

type MetricsGauge struct {
	ID    string
	Value float64
}
type MetricsCounter struct {
	ID    string
	Value int64
}
type Config struct {
	ADDRESS string `mapstructure:"ADDRESS"`
}

func LoadConfig() (config Config, err error) {
	viper.SetDefault("ADDRESS", ":8080")
	viper.AutomaticEnv()
	err = viper.Unmarshal(&config)
	return
}

func main() {
	config, _ := LoadConfig()

	GaugeMetric.metric = make(map[string]float64)
	CounterMetric.metric = make(map[string]int64)

	r := chi.NewRouter()
	r.Use(middleware.Compress(5))
	r.Route("/", func(r chi.Router) {
		r.Get("/", listMetrics)
		r.Post("/{operation}/", func(w http.ResponseWriter, r *http.Request) {
			operation := chi.URLParam(r, "operation")

			if operation != "update" {
				w.WriteHeader(404)
			} else if operation != "value" {
				w.WriteHeader(404)
			}

		})
		r.Post("/update/{metricType}/*", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(404)

		})
		r.Post("/update", receiveMetricJSON)
		r.Post("/value", valueOfMetricJSON)
		r.Post("/update/{metricType}/{metricName}/{metricValue}", receiveMetric)
		r.Get("/value/{metricType}/{metricName}", valueOfMetric)
	})

	http.ListenAndServe(config.ADDRESS, r)
}
