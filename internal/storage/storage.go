package storage

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
)

type MetricsJSON struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

//todo: структуру в модуль +  интерфейс хранения
type Metrics struct {
	MuGauge       sync.RWMutex
	GaugeMetric   map[string]float64 `json:"gaugeMetric"`
	MuCounter     sync.RWMutex
	CounterMetric map[string]int64 `json:"counterMetric"`
}

var MetricsInMem Metrics

func (m *Metrics) InitMetrics() {
	//todo: не понимаю пока что тебе не нравится
	m.GaugeMetric = make(map[string]float64)
	m.CounterMetric = make(map[string]int64)
	fmt.Println("INIT METRICS", m)
}

func (m *Metrics) SaveMetrics(saveConfig *SaveConfig) {
	if saveConfig.ToMem == true {
		m.SaveMetricsToMem()
	}
	if saveConfig.ToFile == true && saveConfig.ToFileSync == true {
		//todo: добавить обработку ошибок
		m.SaveToFile(saveConfig.ToFilePath)
	}
}

func (m *Metrics) GetMetrics() {
	m.GetMetricsFromMem()
}

//todo: добавить сохранение метрик по имени?
//чекнули хранение
func (m *Metrics) SaveMetricsToMem() {
	MetricsInMem.MuGauge.Lock()
	MetricsInMem.GaugeMetric = m.GaugeMetric
	MetricsInMem.MuGauge.Unlock()
	MetricsInMem.MuCounter.Lock()
	MetricsInMem.CounterMetric = m.CounterMetric
	MetricsInMem.MuCounter.Unlock()
}
func (m Metrics) GetMetricsFromMem() {
	if len(MetricsInMem.GaugeMetric) > 0 {
		MetricsInMem.MuGauge.Lock()
		m.GaugeMetric = MetricsInMem.GaugeMetric
		MetricsInMem.MuGauge.Unlock()
	}
	if len(MetricsInMem.CounterMetric) > 0 {
		MetricsInMem.MuCounter.Lock()
		m.CounterMetric = MetricsInMem.CounterMetric
		MetricsInMem.MuCounter.Unlock()
	}
}
func (m *Metrics) SaveToFile(filePath string) {
	fileAttr := os.O_CREATE | os.O_TRUNC | os.O_WRONLY
	file, err := os.OpenFile(filePath, fileAttr, 0644)
	if err != nil {
		log.Fatal(err)
	}
	data, err := json.Marshal(&m)
	if err != nil {
		log.Fatal(err)
	}
	_, err = file.Write(data)
	if err != nil {
		log.Fatal(err)
	}
	err = file.Close()
	if err != nil {
		log.Fatal(err)
	}
}
func (m *Metrics) RestoreFromFile(filePath string) {
	byteFile, _ := ioutil.ReadFile(filePath)
	_ = json.Unmarshal([]byte(byteFile), m)
	fmt.Println("restoring from", m)
}

//закончили упражнение
