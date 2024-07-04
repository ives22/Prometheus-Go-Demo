package demo

import (
	"github.com/prometheus/client_golang/prometheus"
)

func DemoForCounter() prometheus.Counter {
	totalRequests := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "The total number of handled HTTP requests.",
	})

	for i := 0; i <= 10; i++ {
		totalRequests.Inc()
	}

	return totalRequests
}