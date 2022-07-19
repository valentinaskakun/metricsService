package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/valentinaskakun/metricsService/internal/compress"
	"github.com/valentinaskakun/metricsService/internal/config"
	"github.com/valentinaskakun/metricsService/internal/handlers"
	"github.com/valentinaskakun/metricsService/internal/storage"

	"github.com/go-chi/chi/v5"
)

func handleSignal(signal os.Signal) {
	log.Println("* Got:", signal)
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
	//инит конфига
	configRun, err := config.LoadConfigServer()
	if err != nil {
		log.Println(err)
	}
	saveConfigRun, err := config.ParseConfigServer(&configRun)
	if err != nil {
		log.Println(err)
	}
	if configRun.Restore {
		err = metricsRun.RestoreFromFile(configRun.StoreFile)
		if err != nil {
			log.Println(err)
		}
	}
	//если не нужно поддерживать синхронность, создаем тикер -- все равно кривовато
	if !saveConfigRun.ToFileSync {
		storeInterval, err := time.ParseDuration(configRun.StoreInterval)
		if err != nil {
			log.Println(err)
		}
		tickerStore := time.NewTicker(storeInterval)
		{
			go func() {
				for range tickerStore.C {
					metricsRun.GetMetrics(saveConfigRun)
					err := metricsRun.SaveToFile(saveConfigRun.ToFilePath)
					if err != nil {
						log.Print(err)
					}
				}
			}()
		}
	}
	//обработка запросов
	r := chi.NewRouter()
	r.Get("/", handlers.ListMetricsAll(&metricsRun, saveConfigRun))
	r.Route("/update", func(r chi.Router) {
		r.Post("/", handlers.UpdateMetricJSON(&metricsRun, saveConfigRun, configRun.Key))
		r.Post("/{metricType}/{metricName}/{metricValue}", handlers.UpdateMetric(&metricsRun, saveConfigRun))
	})
	r.Post("/updates/", handlers.UpdateMetrics(&metricsRun, saveConfigRun))
	r.Route("/value", func(r chi.Router) {
		r.Post("/", handlers.ListMetricJSON(&metricsRun, saveConfigRun, configRun.Key))
		r.Get("/{metricType}/{metricName}", handlers.ListMetric(&metricsRun, saveConfigRun))
	})
	r.Get("/ping", handlers.Ping(saveConfigRun))
	log.Fatal(http.ListenAndServe(configRun.Address, compress.GzipHandle(r)))
}
