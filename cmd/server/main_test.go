package main

import (
	"github.com/go-resty/resty/v2"
	"testing"
)

var serverAddr = "http://127.0.0.1:8080"

//todo: больше тестов "api"
func TestListMetrics(t *testing.T) {
	//req = serverAddr + "/up"
	client := resty.New()
	http, err := client.R().
		Get(serverAddr)
	if err != nil {
		t.Errorf("URL %v get error %v", serverAddr, err)
	} else if http.StatusCode() != 200 {
		t.Errorf("URL %v code %v", serverAddr, http.StatusCode())
	}
}
func TestPostGetMetrics(t *testing.T) {
	testUpdateLink := serverAddr + "/update/counter/testCount/300"
	clientPost := resty.New()
	clientGet := resty.New()
	http, err := clientPost.R().
		SetHeader("Content-Type", "Content-Type: text/plain").
		Post(testUpdateLink)
	if err != nil {
		t.Errorf("Something went wrong while %v, error %v", testUpdateLink, err)
	} else if http.StatusCode() != 200 {
		t.Errorf("URL %v code %v", testUpdateLink, http.StatusCode())
	}
	http2, err := clientPost.R().
		SetHeader("Content-Type", "Content-Type: text/plain").
		Post(testUpdateLink)
	if err != nil {
		t.Errorf("Something went wrong while %v, error %v", testUpdateLink, err)
	} else if http2.StatusCode() != 200 {
		t.Errorf("URL %v code %v", testUpdateLink, http2.StatusCode())
	}
	testListLink := serverAddr + "/value/counter/testCount"
	http3, err := clientGet.R().
		Get(testListLink)
	if err != nil {
		t.Errorf("URL %v get error %v", testListLink, err)
	} else if http3.StatusCode() != 200 {
		t.Errorf("URL %v code %v", testListLink, http3.StatusCode())
	} else if http3.String() != "600" {
		t.Errorf("Counter results are incorrect, body: %v", http3.String())
	}
}
