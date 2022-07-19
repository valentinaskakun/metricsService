package config

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"os"

	"github.com/caarlos0/env/v6"
	"github.com/rs/zerolog"

	"github.com/valentinaskakun/metricsService/internal/storage"
)

type ConfServer struct {
	Address       string `env:"ADDRESS"`
	StoreInterval string `env:"STORE_INTERVAL"`
	StoreFile     string `env:"STORE_FILE"`
	Key           string `env:"KEY"`
	Database      string `env:"DATABASE_DSN"`
	Restore       bool   `env:"RESTORE"`
}

func LoadConfigServer() (config ConfServer, err error) {
	log := zerolog.New(os.Stdout)
	flag.StringVar(&config.Address, "a", ":8080", "")
	flag.StringVar(&config.StoreInterval, "i", "300s", "")
	flag.StringVar(&config.StoreFile, "f", "/tmp/devops-metrics-db.json", "")
	flag.StringVar(&config.Database, "d", "", "")
	flag.StringVar(&config.Key, "k", "", "")
	flag.BoolVar(&config.Restore, "r", true, "")
	flag.Parse()
	err = env.Parse(&config)
	if err != nil {
		log.Warn().Msg(err.Error())
	}
	return
}

func ParseConfigServer(configIn *ConfServer) (configOut *storage.SaveConfig, err error) {
	processConfig := storage.SaveConfig{}
	processConfig.ToMem = true
	if processConfig.ToMem {
		processConfig.MetricsInMem.InitMetrics()
	}
	if configIn.StoreFile != "" {
		processConfig.ToFile = true
		processConfig.ToFilePath = configIn.StoreFile
	}
	if configIn.Database != "" {
		processConfig.ToDatabase = true
		processConfig.ToDatabaseDSN = configIn.Database
	}
	if processConfig.ToDatabase {
		err := storage.InitTables(&processConfig)
		if err != nil {
			return configOut, err
		}
	}
	if configIn.StoreInterval == "0" {
		processConfig.ToFileSync = true
	}
	configOut = &processConfig
	return configOut, err
}

type ConfAgent struct {
	Address        string `env:"ADDRESS"`
	ReportInterval string `env:"REPORT_INTERVAL"`
	PollInterval   string `env:"POLL_INTERVAL"`
	Key            string `env:"KEY"`
	Proto          string
}

func LoadConfigAgent() (config ConfAgent, err error) {
	log := zerolog.New(os.Stdout)
	config.Proto = "http://"
	flag.StringVar(&config.Address, "a", "localhost:8080", "")
	flag.StringVar(&config.ReportInterval, "r", "10s", "")
	flag.StringVar(&config.PollInterval, "p", "2s", "")
	flag.StringVar(&config.Key, "k", "", "")
	flag.Parse()
	err = env.Parse(&config)
	if err != nil {
		log.Warn().Msg(err.Error())
	}
	return
}

func Hash(msg string, key string) (hash string) {
	src := []byte(msg)
	h := hmac.New(sha256.New, []byte(key))
	h.Write(src)
	hash = hex.EncodeToString(h.Sum(nil))
	return
}
