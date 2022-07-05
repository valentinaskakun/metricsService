package main

import (
	"fmt"
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
	//todo: перенести в config? (в т.ч. inittables)
	var saveConfigRun storage.SaveConfig
	configRun, _ := config.LoadConfigServer()
	saveConfigRun.ToMem = true
	if saveConfigRun.ToMem {
		saveConfigRun.MetricsInMem.InitMetrics()
	}
	if configRun.StoreFile != "" {
		saveConfigRun.ToFile = true
		saveConfigRun.ToFilePath = configRun.StoreFile
	}
	if configRun.Database != "" {
		saveConfigRun.ToDatabase = true
		saveConfigRun.ToDatabaseDSN = configRun.Database
	}
	if configRun.StoreInterval == "0" {
		saveConfigRun.ToFileSync = true
	}
	if configRun.Restore {
		//todo: добавить ошибки на случай отсутствия файла
		metricsRun.RestoreFromFile(configRun.StoreFile)
	}
	if saveConfigRun.ToDatabase {
		err := storage.InitTables(&saveConfigRun)
		if err != nil {
			fmt.Println(err)
		}
	}
	///////TEST
	saveConfigRun.ToDatabase = true
	saveConfigRun.ToDatabaseDSN = "postgres://postgres:postgrespw@localhost:55000"
	storage.InitTables(&saveConfigRun)
	///////TEST
	//если не нужно поддерживать синхронность, создаем тикер, только почему так криво
	if !saveConfigRun.ToFileSync {
		storeInterval, _ := time.ParseDuration(configRun.StoreInterval)
		tickerStore := time.NewTicker(storeInterval)
		{
			go func() {
				for range tickerStore.C {
					metricsRun.GetMetrics(&saveConfigRun)
					metricsRun.SaveToFile(saveConfigRun.ToFilePath)
				}
			}()
		}
	}
	//обработка запросов
	r := chi.NewRouter()
	r.Get("/", handlers.ListMetricsAll(&metricsRun, &saveConfigRun))
	r.Route("/update", func(r chi.Router) {
		r.Post("/", handlers.UpdateMetricJSON(&metricsRun, &saveConfigRun, configRun.Key))
		r.Post("/{metricType}/{metricName}/{metricValue}", handlers.UpdateMetric(&metricsRun, &saveConfigRun))
	})
	r.Post("/updates", handlers.UpdateMetrics(&saveConfigRun))
	r.Route("/value", func(r chi.Router) {
		r.Post("/", handlers.ListMetricJSON(&metricsRun, &saveConfigRun, configRun.Key))
		r.Get("/{metricType}/{metricName}", handlers.ListMetric(&metricsRun, &saveConfigRun))
	})
	r.Get("/ping", handlers.Ping(&saveConfigRun))
	log.Fatal(http.ListenAndServe(configRun.Address, compress.GzipHandle(r)))
}
