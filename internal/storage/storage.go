package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"

	_ "github.com/jackc/pgx/stdlib"
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

func PingDatabase(config *SaveConfig) (err error) {
	db, err := sql.Open("pgx", config.ToDatabaseDSN)
	if err != nil {
		return err
	} else {
		defer db.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		if err = db.PingContext(ctx); err != nil {
			return err
			// to err log
			//fmt.Println("err ping", err)
		}
	}
	return
}

func InitTables(config *SaveConfig) (err error) {
	//как-то надо покрасивее сделать, через структуру видимо
	metricsTable := `
CREATE TABLE IF NOT EXISTS metrics (
  id           TEXT UNIQUE,
  mtype 	  TEXT,
  delta		   BIGINT,
  value        DOUBLE PRECISION
);`
	db, err := sql.Open("pgx", config.ToDatabaseDSN)
	if err != nil {
		return err
	} else {
		defer db.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		_, err := db.ExecContext(ctx, metricsTable)
		if err != nil {
			return err
			// to err log
			//fmt.Println("err ping", err)
		}
	}
	return
}

func UpdateRow(config *SaveConfig, metricsJSON MetricsJSON) (err error) {
	sqlQuery := ""
	//if metricsJSON.MType == "gauge" {
	//	sqlQuery = `INSERT INTO metrics(
	//				id,
	//				type,
	//				value
	//				)
	//				VALUES($1, $2, $3)
	//				ON CONFLICT (id) DO UPDATE
	//				SET value=$3;`
	//} else if metricsJSON.MType == "counter" {
	sqlQuery = `INSERT INTO metrics(
					id,
					mtype,
					delta,
					value
					)
					VALUES($1, $2, $3, $4)
					ON CONFLICT (id) DO UPDATE
					SET delta=metrics.delta+$3, value=$4;`
	db, err := sql.Open("pgx", config.ToDatabaseDSN)
	if err != nil {
		return err
	} else {
		defer db.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_, err = db.ExecContext(ctx, sqlQuery, metricsJSON.ID, metricsJSON.MType, metricsJSON.Delta, metricsJSON.Value)
		//_, err = db.NamedExec(sqlQuery, metricsJSON)
		if err != nil {
			return err
			// to err log
			//fmt.Println("err ping", err)
		}
	}
	return
}
