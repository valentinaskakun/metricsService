package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/valentinaskakun/metricsService/internal/config"
	"github.com/valentinaskakun/metricsService/internal/storage"
	"io/ioutil"
	"net/http"
	"strconv"
)

//todo: добавить интерфейсы для хэндлеров/метод сет?
func ListMetricsAll(metricsRun *storage.Metrics) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		metricsRun.GetMetrics()
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintln(w, "METRICS GAUGE:")
		//todo: нужно ли добавлять RLock
		for key, value := range metricsRun.GaugeMetric {
			fmt.Fprintln(w, key, value)
		}
		fmt.Fprintln(w, "METRICS COUNTER:")
		for key, value := range metricsRun.CounterMetric {
			fmt.Fprintln(w, key, value)
		}
	}
}
func ListMetric(metricsRun *storage.Metrics) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		metricsRun.GetMetrics()
		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")
		if metricType == "gauge" {
			if val, ok := metricsRun.GaugeMetric[metricName]; ok {
				fmt.Fprintln(w, val)
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		} else if metricType == "counter" {
			if val, ok := metricsRun.CounterMetric[metricName]; ok {
				fmt.Fprintln(w, val)
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		} else {
			w.WriteHeader(http.StatusNotImplemented)
		}
	}
}

func ListMetricJSON(metricsRun *storage.Metrics) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		metricsRun.GetMetrics()
		w.Header().Set("Content-Type", "application/json")
		metricReq, metricRes := storage.MetricsJSON{}, storage.MetricsJSON{}
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}
		if err := json.Unmarshal(body, &metricReq); err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}
		if metricReq.MType == "gauge" {
			if _, ok := metricsRun.GaugeMetric[metricReq.ID]; ok {
				metricRes.ID, metricRes.MType, metricRes.Delta = metricReq.ID, metricReq.MType, metricReq.Delta
				valueRes := metricsRun.GaugeMetric[metricReq.ID]
				metricRes.Value = &valueRes
			} else {
				w.WriteHeader(http.StatusNotFound)
				return
			}
		} else if metricReq.MType == "counter" {
			if _, ok := metricsRun.CounterMetric[metricReq.ID]; ok {
				metricRes.ID, metricRes.MType, metricRes.Value = metricReq.ID, metricReq.MType, metricReq.Value
				valueRes := metricsRun.CounterMetric[metricReq.ID]
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
}

func UpdateMetric(metricsRun *storage.Metrics, saveConfig *config.SaveConfig) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		metricsRun.GetMetrics()
		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")
		metricValue := chi.URLParam(r, "metricValue")
		if metricType == "gauge" {
			valParsed, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
			} else {
				metricsRun.MuGauge.Lock()
				metricsRun.GaugeMetric[metricName] = valParsed
				metricsRun.MuGauge.Unlock()
			}
		} else if metricType == "counter" {
			valParsed, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
			} else {
				metricsRun.MuCounter.Lock()
				metricsRun.CounterMetric[metricName] += valParsed
				metricsRun.MuCounter.Unlock()
			}
		} else {
			w.WriteHeader(http.StatusNotImplemented)
		}
		metricsRun.SaveMetrics(saveConfig)
	}
}

func UpdateMetricJSON(metricsRun *storage.Metrics, saveConfig *config.SaveConfig) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		metricsRun.GetMetrics()
		//fmt.Println("MetricsRun result before make", metricsRun.GaugeMetric)
		//metricsRun.GaugeMetric = make(map[string]float64)
		//metricsRun.CounterMetric = make(map[string]int64)
		//fmt.Println("MetricsRun result", metricsRun)
		metricReq := storage.MetricsJSON{}
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}
		if err := json.Unmarshal(body, &metricReq); err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}
		if metricReq.MType == "gauge" {
			metricsRun.MuGauge.Lock()
			metricsRun.GaugeMetric[metricReq.ID] = *metricReq.Value
			metricsRun.MuGauge.Unlock()
		} else if metricReq.MType == "counter" {
			metricsRun.MuCounter.Lock()
			metricsRun.CounterMetric[metricReq.ID] += *metricReq.Delta
			metricsRun.MuCounter.Unlock()
		} else {
			w.WriteHeader(http.StatusNotImplemented)
		}
		metricsRun.SaveMetrics(saveConfig)
		w.WriteHeader(http.StatusOK)
		resBody, _ := json.Marshal("{}")
		w.Write(resBody)
	}
}
