package storage

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"sync"
)

type SaveConfig struct {
	ToMem         bool
	MetricsInMem  Metrics
	ToFile        bool
	ToFilePath    string
	ToFileSync    bool
	ToDatabase    bool
	ToDatabaseDSN string
}

type MetricsJSON struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Hash  string   `json:"hash,omitempty"`  // значение хеш-функции
}

type Metrics struct {
	MuGauge       sync.RWMutex
	GaugeMetric   map[string]float64 `json:"gaugeMetric"`
	MuCounter     sync.RWMutex
	CounterMetric map[string]int64 `json:"counterMetric"`
}

func (m *Metrics) InitMetrics() {
	//todo: разобраться с инициализацией
	m.GaugeMetric = make(map[string]float64)
	m.CounterMetric = make(map[string]int64)
}

func (m *Metrics) SaveMetrics(saveConfig *SaveConfig) {
	if saveConfig.ToMem {
		m.SaveMetricsToMem(&saveConfig.MetricsInMem)
	}
	if saveConfig.ToFile && saveConfig.ToFileSync {
		//todo: добавить обработку ошибок
		m.SaveToFile(saveConfig.ToFilePath)
	}
}

func (m *Metrics) GetMetrics(saveConfig *SaveConfig) {
	m.GetMetricsFromMem(&saveConfig.MetricsInMem)
}

//todo: добавить сохранение метрик по имени?
func (m *Metrics) SaveMetricsToMem(metricsInMem *Metrics) {
	metricsInMem.MuGauge.Lock()
	metricsInMem.GaugeMetric = m.GaugeMetric
	metricsInMem.MuGauge.Unlock()
	metricsInMem.MuCounter.Lock()
	metricsInMem.CounterMetric = m.CounterMetric
	metricsInMem.MuCounter.Unlock()
}
func (m *Metrics) GetMetricsFromMem(metricsInMem *Metrics) {
	if len(metricsInMem.GaugeMetric) > 0 {
		metricsInMem.MuGauge.Lock()
		m.GaugeMetric = metricsInMem.GaugeMetric
		metricsInMem.MuGauge.Unlock()
	}
	if len(metricsInMem.CounterMetric) > 0 {
		metricsInMem.MuCounter.Lock()
		m.CounterMetric = metricsInMem.CounterMetric
		metricsInMem.MuCounter.Unlock()
	}
}
func (m *Metrics) SaveToFile(filePath string) {
	fileAttr := os.O_CREATE | os.O_TRUNC | os.O_WRONLY
	file, err := os.OpenFile(filePath, fileAttr, 0644)
	if err != nil {
		log.Println(err)
	}
	data, err := json.Marshal(&m)
	if err != nil {
		log.Println(err)
	}
	_, err = file.Write(data)
	if err != nil {
		log.Println(err)
	}
	err = file.Close()
	if err != nil {
		log.Fatal(err)
	}
}
func (m *Metrics) RestoreFromFile(filePath string) {
	byteFile, _ := ioutil.ReadFile(filePath)
	_ = json.Unmarshal(byteFile, m)
}
