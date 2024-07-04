package demo

import "github.com/prometheus/client_golang/prometheus"

func DemoForGauge() prometheus.Gauge {
	queueLength := prometheus.NewGauge(prometheus.GaugeOpts{
		// Namespace,Subsystem, Name 会拼接成指标的名称：system_web_queue_length
		// 其中Name是必填参数
		Namespace: "system",
		Subsystem: "web",
		Name:      "queue_length",
		// 指标的描述信息
		Help: "The number of items in the queue",
		// 指标的标签
		ConstLabels: map[string]string{
			"module": "http-server",
		},
	})

	queueLength.Inc()   // +1: gauge增加1
	queueLength.Dec()   // -1: gauge减少1
	queueLength.Add(20) // 增加20个增量
	queueLength.Sub(12) // 减少12个

	return queueLength
}