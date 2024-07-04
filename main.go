package main

import (
	"Prometheus-Go-Demo/collector"
	"Prometheus-Go-Demo/collectorv1"
	"Prometheus-Go-Demo/demo"
	"flag"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
)

var (

	// 命令行参数
	listenAddr       = flag.String("web.listen-port", "8090", "An port to listen on for web interface and telemetry.")
	metricsPath      = flag.String("web.telemetry-path", "/metrics", "A path under which to expose metrics.")
	metricsNamespace = flag.String("metric.namespace", "app", "Prometheus metrics namespace, as the prefix of metrics name")
)

func main() {

	// 解析命令行参数
	flag.Parse()

	//	创建一个自定义的注册表
	registry := prometheus.NewRegistry()

	dGauge := demo.DemoForGauge()
	dCounter := demo.DemoForCounter()
	dSummary := demo.DemoForSummary()
	dHistogram := demo.DemoForHistogram()

	registry.MustRegister(dGauge)
	registry.MustRegister(dCounter)
	registry.MustRegister(dSummary)
	registry.MustRegister(dHistogram)

	// #################################################
	memUsageCollector := collector.NewMemUsageCollectorDemo()
	apiRequestCollector := collector.NewApiRequestCounterDemo()
	//cpuUsageCollector := collector.NewCpuUsageSummaryCollectorDemo()

	// #################################################
	// 动态添加标签的 collector 版本

	testLables := map[string]string{
		"ip":        "127.0.0.1",
		"host_name": "test_cvm",
	}

	memUsageCollectorV1 := collectorv1.NewMemUsageCollectorDemo(*metricsNamespace, testLables)
	apiRequestCollectorV1 := collectorv1.NewApiRequestCounterDemo(*metricsNamespace, testLables)

	// 注册memUsageCollector实例到注册表registry中
	registry.MustRegister(memUsageCollector)
	// 注册memUsageCollectorV1实例到注册表registry中
	registry.MustRegister(memUsageCollectorV1)

	// 注册apiRequestCollector实例到注册表registry中
	registry.MustRegister(apiRequestCollector)
	// 注册apiRequestCollectorV1实例到注册表registry中
	registry.MustRegister(apiRequestCollectorV1)

	//registry.MustRegister(cpuUsageCollector)

	// 创建一个api接口，用于模拟测试API请求处理函数
	http.HandleFunc("/api", func(writer http.ResponseWriter, request *http.Request) {
		apiRequestCollector.IncrementRequestCount()
		apiRequestCollectorV1.IncrementRequestCount()
		writer.Write([]byte("API请求处理成功"))
	})

	// 处理根页面
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte(`<html>
	            <head><title>A Prometheus Exporter</title></head>
	            <body>
	            <h1>Prometheus Exporter</h1>
	            <p><a href='/metrics'>Metrics</a></p>
	            </body>
	            </html>`))
	})

	// 设置HTTP服务器以处理Prometheus指标的HTTP请求
	http.Handle(*metricsPath, promhttp.HandlerFor(registry, promhttp.HandlerOpts{Registry: registry}))

	// 记录启动日志并启动HTTP服务器监听
	log.Printf("Starting Server at http://localhost:%s%s", *listenAddr, *metricsPath)
	log.Fatal(http.ListenAndServe(":"+*listenAddr, nil))
}