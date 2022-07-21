package metricssend

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"

	"github.com/valentinaskakun/metricsService/internal/config"
	"github.com/valentinaskakun/metricsService/internal/storage"
)

func SendMetricJSON(metricsToSend *storage.Metrics, serverToSendLink string, configRun *config.ConfAgent) {
	log := zerolog.New(os.Stdout)
	if metricsToSend.CounterMetric["PollCount"] != 0 {
		urlStr, err := url.Parse(serverToSendLink)
		if err != nil {
			log.Warn().Msg(err.Error())
			return
		}
		urlStr.Path = path.Join(urlStr.Path, "update")
		client := resty.New()
		client.R().
			SetHeader("Content-Type", "Content-Type: application/json")
		for key, value := range metricsToSend.GaugeMetric {
			metricToSend, err := json.Marshal(storage.MetricsJSON{ID: key, MType: "gauge", Value: &value})
			if err != nil {
				log.Warn().Msg(err.Error())
				return
			}
			if len(configRun.Key) > 0 {
				//todo: переделать функцию хэш с нормальными аргументами
				hashValue := config.Hash(fmt.Sprintf("%s:gauge:%f", key, value), configRun.Key)
				metricToSend, err = json.Marshal(storage.MetricsJSON{ID: key, MType: "gauge", Value: &value, Hash: hashValue})
				if err != nil {
					log.Warn().Msg(err.Error())
					return
				}
			}
			_, err = client.R().
				SetBody(metricToSend).
				Post(urlStr.String())
			if err != nil {
				log.Warn().Msg(err.Error())
				return
			}
		}
		for key, value := range metricsToSend.CounterMetric {
			metricToSend, err := json.Marshal(storage.MetricsJSON{ID: key, MType: "counter", Delta: &value})
			if err != nil {
				log.Warn().Msg(err.Error())
				return
			}
			if len(configRun.Key) > 0 {
				hashValue := config.Hash(fmt.Sprintf("%s:counter:%d", key, value), configRun.Key)
				metricToSend, err = json.Marshal(storage.MetricsJSON{ID: key, MType: "counter", Delta: &value, Hash: hashValue})
				if err != nil {
					log.Warn().Msg(err.Error())
					return
				}
			}
			_, err = client.R().
				SetBody(metricToSend).
				Post(urlStr.String())
			if err != nil {
				log.Warn().Msg(err.Error())
				return
			}
		}
	} else {
		fmt.Println("ERROR: Something went wrong while sendingMetricJSON")
	}
}
func SendMetricsBatch(metricsToSend *storage.Metrics, serverToSendLink string) {
	log := zerolog.New(os.Stdout)
	var metricsBatch []storage.MetricsJSON
	if metricsToSend.CounterMetric["PollCount"] == 0 {
		log.Info().Msg("ERROR: PollCount is Null")
		return
	} else {
		urlStr, err := url.Parse(serverToSendLink)
		if err != nil {
			log.Warn().Msg(err.Error())
		}
		urlStr.Path = path.Join(urlStr.Path, "updates")
		client := resty.New()
		client.R().
			SetHeader("Content-Type", "Content-Type: application/json")
		for key, value := range metricsToSend.GaugeMetric {
			newVal := value
			metricToSend := storage.MetricsJSON{ID: key, MType: "gauge", Value: &newVal}
			metricsBatch = append(metricsBatch, metricToSend)
		}
		for key, value := range metricsToSend.CounterMetric {
			newVal := value
			metricToSend := storage.MetricsJSON{ID: key, MType: "counter", Delta: &newVal}
			metricsBatch = append(metricsBatch, metricToSend)
		}
		if len(metricsBatch) > 0 {
			metricsPrepared, err := json.Marshal(metricsBatch)
			if err != nil {
				log.Warn().Msg(err.Error())
				return
			}
			_, err = client.R().
				SetBody(metricsPrepared).
				Post(urlStr.String() + "/")
			if err != nil {
				log.Warn().Msg(err.Error())
				return
			}
		} else {
			return
		}

	}
}
