package demo

import "github.com/prometheus/client_golang/prometheus"

func DemoForSummary() prometheus.Summary {
	requestDurations := prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "http_request_duration_for_summary_seconds",
		Help: "A summary of the HTTP request durations in seconds.",
		Objectives: map[float64]float64{
			0.5:  0.05,  // 第50个百分位数，最大绝对误差为0.05。
			0.9:  0.01,  // 第90个百分位数，最大绝对误差为0.01。
			0.99: 0.001, // 第90个百分位数，最大绝对误差为0.001。
		},
	},
	)

	// 添加值
	for _, v := range []float64{0.01, 0.02, 0.3, 0.4, 0.6, 0.7, 5.5, 11} {
		requestDurations.Observe(v)
	}

	return requestDurations
}