package storage

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"
	"time"
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
	TimeMetric    time.Time
}

var MetricsInMem Metrics

func (m *Metrics) InitMetrics() {
	//todo: не понимаю пока что тебе не нравится
	tempMetrics := new(Metrics)
	m = tempMetrics
	m.GaugeMetric = make(map[string]float64)
	m.CounterMetric = make(map[string]int64)
}

func (m *Metrics) SaveMetrics(filePath string) {
	m.SaveMetricsToMem()
	m.SaveToFile(filePath)
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
	m.InitMetrics()
	MetricsInMem.MuGauge.Lock()
	m.GaugeMetric = MetricsInMem.GaugeMetric
	MetricsInMem.MuGauge.Unlock()
	MetricsInMem.MuCounter.Lock()
	m.CounterMetric = MetricsInMem.CounterMetric
	MetricsInMem.MuCounter.Unlock()
}
func (m *Metrics) SaveToFile(filePath string) {
	file, _ := json.Marshal(&m)
	err := ioutil.WriteFile(filePath, file, 0644)
	fmt.Println(err)
}
func (m *Metrics) RestoreFromFile(filePath string) {
	byteFile, _ := ioutil.ReadFile(filePath)
	data := Metrics{}
	_ = json.Unmarshal([]byte(byteFile), &data)
	fmt.Println(data)
}

//закончили упражнение
