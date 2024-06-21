package collectorv1

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"sync/atomic"
)

type ApiRequestCounterDemo struct {
	ApiRequestDesc *prometheus.Desc
	LabelValues    []string // 动态标签的值
	requestCount   uint64
	prefix         string // 指标名前缀
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

func NewApiRequestCounterDemo(prefix string, labels map[string]string) *ApiRequestCounterDemo {
	var (
		keys   []string
		values []string
	)

	for k, v := range labels {
		keys = append(keys, k)
		values = append(values, v)
	}

	fqName := fmt.Sprintf("%s_api_request_count_total_v1", prefix)
	return &ApiRequestCounterDemo{
		ApiRequestDesc: prometheus.NewDesc(
			fqName,
			"API请求总数v1",
			keys,
			prometheus.Labels{"test": "true"},
		),
		LabelValues: values,
		prefix:      prefix,
	}
}

// IncrementRequestCount 用于增加API请求计数
func (a *ApiRequestCounterDemo) IncrementRequestCount() {
	atomic.AddUint64(&a.requestCount, 1)
}