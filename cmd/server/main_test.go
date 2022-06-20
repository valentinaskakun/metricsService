package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/valentinaskakun/metricsService/internal/handlers"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

var serverAddr = "http://127.0.0.1:8080"

//todo: поменять ее на импортированную из репо
func testRequest(t *testing.T, ts *httptest.Server, method, path string, body io.Reader) (*http.Response, string) {
	req, _ := http.NewRequest(method, ts.URL+path, body)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}
	defer resp.Body.Close()

	return resp, string(respBody)
}

//todo: больше тестов "api"
func TestListMetrics(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	h := http.HandlerFunc(handlers.ListMetricsAll)
	h.ServeHTTP(w, req)
	res := w.Result()
	defer res.Body.Close()
	if res.StatusCode != 200 {
		t.Errorf("URL %v code %v", serverAddr, res.StatusCode)
	}
}
func TestPostGetMetrics(t *testing.T) {
	r := chi.NewRouter()
	r.Post("/update/{metricType}/{metricName}/{metricValue}", handlers.UpdateMetric)
	r.Get("/value/{metricType}/{metricName}", handlers.ListMetric)
	ts := httptest.NewServer(r)
	defer ts.Close()
	testUpdateLink := "/update/counter/testCount/300"
	req, _ := http.NewRequest("POST", ts.URL+testUpdateLink, nil)
	fmt.Println(req)
	req.Header.Set("Content-Type", "Content-Type: text/plain")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Errorf("Url %v ERROR: %v", testUpdateLink, err)
	} else if res.StatusCode != 200 {
		t.Errorf("URL %v code %v", testUpdateLink, res.StatusCode)
	}
	res.Body.Close()
}
