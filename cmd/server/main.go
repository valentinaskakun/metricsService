package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
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
}

func valueOfMetricJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	metricReq := Metrics{}
	metricRes := Metrics{}
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
func handleSignal(signal os.Signal) {
	fmt.Println("* Got:", signal)
	os.Exit(-1)
}
func main() {
	sigs := make(chan os.Signal, 4)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	go func() {
		for {
			sig := <-sigs
			handleSignal(sig)
		}
	}()
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
