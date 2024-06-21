package collectorv1

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/mem"
)

// MemUsageCollectorDemo 是一个自定义的Prometheus Collector，用于收集系统内存使用量
type MemUsageCollectorDemo struct {
	MemUsageDesc *prometheus.Desc // 描述指标的元数据，包括名称、帮助信息、标签等
	LabelValues  []string         // 动态标签的值
	MemUsage     float64          // 实际监控的内存使用量值
	prefix       string           // 指标名前缀
}

// Describe 实现了Prometheus的Collector接口，用于发送描述符到Prometheus的描述符通道
func (m *MemUsageCollectorDemo) Describe(ch chan<- *prometheus.Desc) {
	// 将内存使用量的描述符发送到描述符通道
	ch <- m.MemUsageDesc
}

// Collect 实现了Prometheus的Collector接口，用于收集指标并发送到Prometheus的指标通道
func (m *MemUsageCollectorDemo) Collect(ch chan<- prometheus.Metric) {
	// 更新内存使用量
	m.updateMemUsage()
	// 创建新的指标并发送到指标通道
	ch <- prometheus.MustNewConstMetric(
		m.MemUsageDesc,        // 描述符
		prometheus.GaugeValue, // 指标类型
		m.MemUsage,            // 指标值
		m.LabelValues...,      // 动态标签的值
	)
}

// NewMemUsageCollectorDemo 是 MemUsageCollectorDemo 的构造函数，返回一个新的实例   prefix 表示指标的前缀，labels 是动态标签
func NewMemUsageCollectorDemo(prefix string, labels map[string]string) *MemUsageCollectorDemo {
	// 处理动态标签
	var (
		keys   []string
		values []string
	)
	for k, v := range labels {
		keys = append(keys, k)
		values = append(values, v)
	}

	fqName := fmt.Sprintf("%s_memory_usage_bytes_v1", prefix)

	return &MemUsageCollectorDemo{
		// 创建一个新的描述符，用于描述内存使用量的指标
		MemUsageDesc: prometheus.NewDesc(
			fqName,                             // 指标名称
			"系统内存使用量v1",                        // 指标帮助信息
			keys,                               // 动态标签
			prometheus.Labels{"module": "mem"}, // 静态标签
		),
		// 设置动态标签的值
		LabelValues: values,
		prefix:      prefix,
	}
}

// updateMemUsage 更新 MemUsageCollectorDemo 实例的内存使用量
func (m *MemUsageCollectorDemo) updateMemUsage() {
	// 使用gopsutil库获取系统内存信息
	memInfo, _ := mem.VirtualMemory()
	// 将可用内存量转换为float64并更新到MemUsage字段
	m.MemUsage = float64(memInfo.Available)
}