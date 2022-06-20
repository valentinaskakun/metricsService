package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/spf13/viper"
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

type ServerConfig struct {
	ADDRESS        string `mapstructure:"ADDRESS"`
	STORE_INTERVAL string `mapstructure:"STORE_INTERVAL"`
	STORE_FILE     string `mapstructure:"STORE_FILE"`
	RESTORE        bool   `mapstructure:"RESTORE"`
}

var SaveConfigRun storage.SaveConfig

func loadConfig() (config ServerConfig, err error) {
	viper.SetDefault("ADDRESS", "localhost:8080")
	viper.SetDefault("STORE_INTERVAL", "10s")
	//viper.SetDefault("STORE_FILE", "/tmp/devops-metrics-db.json")
	viper.SetDefault("STORE_FILE", "D:\\inn\\go\\yapraktikum\\metricsService\\cmd\\server\\test.json")
	viper.SetDefault("RESTORE", "true")
	viper.AutomaticEnv()
	err = viper.Unmarshal(&config)
	SaveConfigRun.ToMem = true
	if config.STORE_FILE != "" {
		SaveConfigRun.ToFile = true
		SaveConfigRun.ToFilePath = config.STORE_FILE
	}
	if config.STORE_INTERVAL == "0" {
		SaveConfigRun.ToFileSync = true
	}
	return
}

func handleSignal(signal os.Signal) {
	fmt.Println("* Got:", signal)
	os.Exit(-1)
}

func main() {
	var metricsRun storage.Metrics
	metricsRun.InitMetrics()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	go func() {
		for {
			sig := <-sigs
			handleSignal(sig)
		}
	}()
	configRun, _ := loadConfig()
	if configRun.RESTORE {
		metricsRun.RestoreFromFile(configRun.STORE_FILE)
	}
	fmt.Println("PRINT AFTER RESTORING", metricsRun)
	if !SaveConfigRun.ToFileSync {
		storeInterval, _ := time.ParseDuration(configRun.STORE_INTERVAL)
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
	log.Fatal(http.ListenAndServe(configRun.ADDRESS, r))
}
