package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/valentinaskakun/metricsService/internal/config"
	"time"
	//todo: ?
	"github.com/valentinaskakun/metricsService/internal/handlers"
	"github.com/valentinaskakun/metricsService/internal/storage"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func handleSignal(signal os.Signal) {
	fmt.Println("* Got:", signal)
	os.Exit(-1)
}

func main() {
	//обработка сигналов
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	go func() {
		for {
			sig := <-sigs
			handleSignal(sig)
		}
	}()
	//инициализировали структуру, в которой работаем с метриками
	var metricsRun storage.Metrics
	metricsRun.InitMetrics()
	//парс конфига
	var SaveConfigRun config.SaveConfig
	configRun, _ := config.LoadConfigServer()
	SaveConfigRun.ToMem = true
	if configRun.StoreFile != "" {
		SaveConfigRun.ToFile = true
		SaveConfigRun.ToFilePath = configRun.StoreFile
	}
	if configRun.StoreInterval == "0" {
		SaveConfigRun.ToFileSync = true
	}
	if configRun.Restore {
		metricsRun.RestoreFromFile(configRun.StoreFile)
	}
	//если не нужно поддерживать синхронность, создаем тикер, только почему так криво
	if !SaveConfigRun.ToFileSync {
		storeInterval, _ := time.ParseDuration(configRun.StoreInterval)
		tickerStore := time.NewTicker(storeInterval)
		{
			go func() {
				for range tickerStore.C {
					metricsRun.GetMetrics()
					metricsRun.SaveToFile(SaveConfigRun.ToFilePath)
				}
			}()
		}
	}
	//обработка запросов
	r := chi.NewRouter()
	r.Get("/", handlers.ListMetricsAll(&metricsRun))
	r.Route("/update", func(r chi.Router) {
		r.Post("/", handlers.UpdateMetricJSON(&metricsRun, &SaveConfigRun))
		r.Post("/{metricType}/{metricName}/{metricValue}", handlers.UpdateMetric(&metricsRun, &SaveConfigRun))
	})
	r.Route("/value", func(r chi.Router) {
		r.Post("/", handlers.ListMetricJSON(&metricsRun))
		r.Get("/{metricType}/{metricName}", handlers.ListMetric(&metricsRun))
	})
	log.Fatal(http.ListenAndServe(configRun.Address, r))
}
