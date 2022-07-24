package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/counter", Sayhello)
	http.ListenAndServe(":8080", nil)
}
