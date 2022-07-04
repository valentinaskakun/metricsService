package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/valentinaskakun/metricsService/internal/config"
	"github.com/valentinaskakun/metricsService/internal/storage"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/stdlib"
)

//todo: добавить интерфейсы для хэндлеров/метод сет?
func ListMetricsAll(metricsRun *storage.Metrics, saveConfig *storage.SaveConfig) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		metricsRun.GetMetrics(saveConfig)
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
func ListMetric(metricsRun *storage.Metrics, saveConfig *storage.SaveConfig) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		metricsRun.GetMetrics(saveConfig)
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

func ListMetricJSON(metricsRun *storage.Metrics, saveConfig *storage.SaveConfig, useHash string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		metricsRun.GetMetrics(saveConfig)
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
				if len(useHash) > 0 {
					metricRes.Hash = config.Hash(fmt.Sprintf("%s:gauge:%f", metricRes.ID, *metricRes.Value), useHash)
				}
			} else {
				w.WriteHeader(http.StatusNotFound)
				return
			}
		} else if metricReq.MType == "counter" {
			if _, ok := metricsRun.CounterMetric[metricReq.ID]; ok {
				metricRes.ID, metricRes.MType, metricRes.Value = metricReq.ID, metricReq.MType, metricReq.Value
				valueRes := metricsRun.CounterMetric[metricReq.ID]
				metricRes.Delta = &valueRes
				if len(useHash) > 0 {
					metricRes.Hash = config.Hash(fmt.Sprintf("%s:counter:%d", metricRes.ID, *metricRes.Delta), useHash)
				}
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

func UpdateMetric(metricsRun *storage.Metrics, saveConfig *storage.SaveConfig) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		metricsRun.GetMetrics(saveConfig)
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

func UpdateMetricJSON(metricsRun *storage.Metrics, saveConfig *storage.SaveConfig, useHash string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		metricsRun.GetMetrics(saveConfig)
		metricReq := storage.MetricsJSON{}
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}
		if err := json.Unmarshal(body, &metricReq); err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}
		if metricReq.MType == "gauge" {
			if (len(useHash) > 0) && (metricReq.Hash != config.Hash(fmt.Sprintf("%s:gauge:%f", metricReq.ID, *metricReq.Value), useHash)) {
				w.WriteHeader(http.StatusBadRequest)
			} else {
				metricsRun.MuGauge.Lock()
				metricsRun.GaugeMetric[metricReq.ID] = *metricReq.Value
				metricsRun.MuGauge.Unlock()
			}
		} else if metricReq.MType == "counter" {
			if (len(useHash) > 0) && (metricReq.Hash != config.Hash(fmt.Sprintf("%s:counter:%d", metricReq.ID, *metricReq.Delta), useHash)) {
				w.WriteHeader(http.StatusBadRequest)
			} else {
				metricsRun.MuCounter.Lock()
				metricsRun.CounterMetric[metricReq.ID] += *metricReq.Delta
				metricsRun.MuCounter.Unlock()
			}
		} else {
			w.WriteHeader(http.StatusNotImplemented)
		}
		metricsRun.SaveMetrics(saveConfig)
		w.WriteHeader(http.StatusOK)
		resBody, _ := json.Marshal("{}")
		w.Write(resBody)
	}
}
func Ping(saveConfig *storage.SaveConfig) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(saveConfig)
		//saveConfig.ToDatabase = true
		if saveConfig.ToDatabase {
			//todo: вынести логику бд в storage.go
			db, err := sql.Open("pgx", saveConfig.ToDatabaseDSN)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				defer db.Close()
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
				defer cancel()
				if err = db.PingContext(ctx); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					// to err log
					fmt.Println("err ping", err)
				} else {
					w.WriteHeader(http.StatusOK)
				}
			}
			// to err log
			fmt.Println("err", err)
		} else {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "Database DSN isn't set")
		}
	}
}
