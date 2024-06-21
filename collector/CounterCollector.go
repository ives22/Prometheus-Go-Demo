package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"sync/atomic"
)

type ApiRequestCounterDemo struct {
	ApiRequestDesc *prometheus.Desc
	LabelValues    []string // 动态标签的值
	requestCount   uint64
}

func (a *ApiRequestCounterDemo) Describe(ch chan<- *prometheus.Desc) {
	ch <- a.ApiRequestDesc
}

func (a *ApiRequestCounterDemo) Collect(ch chan<- prometheus.Metric) {
	count := atomic.LoadUint64(&a.requestCount)
	ch <- prometheus.MustNewConstMetric(
		a.ApiRequestDesc,
		prometheus.CounterValue,
		float64(count),
		a.LabelValues...,
	)
}

func NewApiRequestCounterDemo() *ApiRequestCounterDemo {
	return &ApiRequestCounterDemo{
		ApiRequestDesc: prometheus.NewDesc(
			"api_request_count_total",
			"API请求总数",
			[]string{"api"},
			prometheus.Labels{"test": "true"},
		),
		LabelValues: []string{"test_api"},
	}
}

// IncrementRequestCount 用于增加API请求计数
func (a *ApiRequestCounterDemo) IncrementRequestCount() {
	atomic.AddUint64(&a.requestCount, 1)
}