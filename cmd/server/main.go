package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Metrics struct {
	gaugeMetric   map[string]float64
	counterMetric map[string]int64
	timeMetric    time.Time
}

var serverToGetAddress = "127.0.0.1:8080"

var MetricsRun Metrics

//var serverToGetProto = "http"
func listMetrics(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "METRICS:")
	for key, value := range MetricsRun.gaugeMetric {
		fmt.Fprintln(w, key, value)
	}
	for key, value := range MetricsRun.counterMetric {
		fmt.Fprintln(w, key, value)
	}

}
func updateMetrics(w http.ResponseWriter, r *http.Request) {
	fmt.Println("ya zdes", r.RequestURI)
	//w.WriteHeader(http.StatusOK)
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")
	metricValue := chi.URLParam(r, "metricValue")
	//fmt.Fprintln(w, "bogdan update:"+chi.URLParam(r, "metricType"))
	if metricType == "gauge" {
		fmt.Println("bogdan gauge:" + chi.URLParam(r, "metricType"))
		valParsed, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			w.WriteHeader(400)
		} else {
			MetricsRun.gaugeMetric[metricName] = valParsed
		}
	} else if metricType == "counter" {
		fmt.Fprintln(w, "bogdan counter:"+chi.URLParam(r, "metricType"))
		valParsed, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			w.WriteHeader(400)
		} else {
			MetricsRun.counterMetric[metricName] = valParsed
		}
	} else {
		fmt.Println("bogdan else:" + chi.URLParam(r, "metricType"))
		w.WriteHeader(501)
	}
	//splitURL := strings.Split(r.RequestURI, "/")
	//if len(splitURL) == 5 {
	//	if splitURL[2] == "gauge" {
	//		metricName := splitURL[3]
	//		metricValue, _ := strconv.ParseFloat(splitURL[4], 64)
	//		MetricsRun.gaugeMetric[metricName] = metricValue
	//	}
	//	if splitURL[2] == "counter" {
	//		metricName := splitURL[3]
	//		metricValue, _ := strconv.ParseInt(splitURL[4], 10, 64)
	//		MetricsRun.counterMetric[metricName] += metricValue
	//	}
	//}
}

func main() {
	MetricsRun.gaugeMetric = make(map[string]float64)
	MetricsRun.counterMetric = make(map[string]int64)
	r := chi.NewRouter()
	r.Post("/", listMetrics)
	r.Route("/update", func(r chi.Router) {
		r.Route("/{metricType}", func(r chi.Router) {
			r.Get("/{metricName}/{metricValue}", updateMetrics)
		})
	})

	//r.Get("/update/{metricType}/{metricName}/{metricValue}", updateMetrics)
	//r.Route("/{carID}", func(r chi.Router) {
	log.Fatal(http.ListenAndServe(serverToGetAddress, r))
}
