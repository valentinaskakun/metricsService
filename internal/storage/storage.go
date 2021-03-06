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
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
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
	ID    string   `json:"id" ,db:"id"`                 // имя метрики
	MType string   `json:"type" ,db:"mtype"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty" ,db:"delta"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty" ,db:"value"` // значение метрики в случае передачи gauge
	Hash  string   `json:"hash,omitempty"`              // значение хеш-функции
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
	//по идее, это должно работать с батчами
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
	log := zerolog.New(os.Stdout)
	db, err := sql.Open("pgx", config.ToDatabaseDSN)
	if err != nil {
		return err
	} else {
		defer db.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		if err = db.PingContext(ctx); err != nil {
			log.Warn().Msg(err.Error())
			return err
		}
	}
	return
}

func InitTables(config *SaveConfig) (err error) {
	log := zerolog.New(os.Stdout)
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
		log.Warn().Msg(err.Error())
		return err
	} else {
		defer db.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		_, err := db.ExecContext(ctx, metricsTable)
		if err != nil {
			log.Warn().Msg(err.Error())
			return err
		}
	}
	return
}

func UpdateRow(config *SaveConfig, metricsJSON *MetricsJSON) (err error) {
	sqlQuery := `INSERT INTO metrics(
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
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()
		_, err = db.ExecContext(ctx, sqlQuery, metricsJSON.ID, metricsJSON.MType, metricsJSON.Delta, metricsJSON.Value)
		if err != nil {
			return err
		}
	}
	return
}

func UpdateBatch(config *SaveConfig, metricsBatch []MetricsJSON) (err error) {
	sqlQuery := `INSERT INTO metrics (id,
				mtype,
				delta,
				value)
	VALUES ($1, $2, $3, $4) ON CONFLICT (id) DO UPDATE
				SET delta=metrics.delta+$3, value=$4;`
	db, err := sql.Open("pgx", config.ToDatabaseDSN)
	if err != nil {
		return err
	} else {
		defer db.Close()
		txn, err := db.Begin()
		if err != nil {
			return errors.Wrap(err, "could not start a new transaction")
		}
		for _, metric := range metricsBatch {
			_, err = txn.Exec(sqlQuery, metric.ID, metric.MType, metric.Delta, metric.Value)
			if err != nil {
				txn.Rollback()
				return errors.Wrap(err, "failed to insert multiple records at once")
			}
		}
		if err := txn.Commit(); err != nil {
			return errors.Wrap(err, "failed to commit transaction")
		}
	}
	return
}
