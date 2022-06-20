package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"log"
)

type SaveConfig struct {
	ToMem      bool
	ToFile     bool
	ToFilePath string
	ToFileSync bool
	ToDatabase bool
}

type ConfServer struct {
	Address       string `env:"ADDRESS"`
	StoreInterval string `env:"STORE_INTERVAL"`
	StoreFile     string `env:"STORE_FILE"`
	Restore       bool   `env:"RESTORE"`
}

func LoadConfigServer() (config ConfServer, err error) {
	flag.StringVar(&config.Address, "a", ":8080", "")
	flag.StringVar(&config.StoreInterval, "i", "300s", "")
	flag.StringVar(&config.StoreFile, "f", "/tmp/devops-metrics-db.json", "")
	flag.BoolVar(&config.Restore, "r", true, "")
	flag.Parse()
	err = env.Parse(&config)
	if err != nil {
		log.Fatal(err)
	}
	return
}

type ConfAgent struct {
	Address        string `env:"ADDRESS"`
	ReportInterval string `env:"REPORT_INTERVAL"`
	PollInterval   string `env:"POLL_INTERVAL"`
}

func LoadConfigAgent() (config ConfAgent, err error) {
	flag.StringVar(&config.Address, "a", "localhost:8080", "")
	flag.StringVar(&config.ReportInterval, "r", "10s", "")
	flag.StringVar(&config.PollInterval, "p", "2s", "")
	flag.Parse()
	err = env.Parse(&config)
	if err != nil {
		log.Fatal(err)
	}
	return
}
