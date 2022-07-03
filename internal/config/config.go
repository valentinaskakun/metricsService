package config

import (
<<<<<<< HEAD
	"crypto/hmac"
	"crypto/sha256"
=======
>>>>>>> b98348e5dd4422ed0ca183164d6020297ad88c8a
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
)

type ConfServer struct {
	Address       string `env:"ADDRESS"`
	StoreInterval string `env:"STORE_INTERVAL"`
	StoreFile     string `env:"STORE_FILE"`
<<<<<<< HEAD
	Key           string `env:"KEY"`
=======
>>>>>>> b98348e5dd4422ed0ca183164d6020297ad88c8a
	Restore       bool   `env:"RESTORE"`
}

func LoadConfigServer() (config ConfServer, err error) {
	flag.StringVar(&config.Address, "a", ":8080", "")
	flag.StringVar(&config.StoreInterval, "i", "300s", "")
	flag.StringVar(&config.StoreFile, "f", "/tmp/devops-metrics-db.json", "")
<<<<<<< HEAD
	flag.StringVar(&config.Key, "k", "", "")
=======
>>>>>>> b98348e5dd4422ed0ca183164d6020297ad88c8a
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
<<<<<<< HEAD
	Key            string `env:"KEY"`
=======
>>>>>>> b98348e5dd4422ed0ca183164d6020297ad88c8a
}

func LoadConfigAgent() (config ConfAgent, err error) {
	flag.StringVar(&config.Address, "a", "localhost:8080", "")
	flag.StringVar(&config.ReportInterval, "r", "10s", "")
	flag.StringVar(&config.PollInterval, "p", "2s", "")
<<<<<<< HEAD
	flag.StringVar(&config.Key, "k", "", "")
=======
>>>>>>> b98348e5dd4422ed0ca183164d6020297ad88c8a
	flag.Parse()
	err = env.Parse(&config)
	if err != nil {
		log.Fatal(err)
	}
	return
}
<<<<<<< HEAD

func Hash(msg string, key string) (hash string, err error) {
	src := []byte(msg)
	h := hmac.New(sha256.New, []byte(key))
	h.Write(src)
	hash = string(h.Sum(nil))
	if err != nil {
		log.Fatal(err)
	}
	return
}
=======
>>>>>>> b98348e5dd4422ed0ca183164d6020297ad88c8a
