package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/mem"
)

// MemUsageCollectorDemo 是一个自定义的Prometheus Collector，用于收集系统内存使用量
type MemUsageCollectorDemo struct {
	MemUsageDesc *prometheus.Desc // 描述指标的元数据，包括名称、帮助信息、标签等
	LabelValues  []string         // 动态标签的值
	MemUsage     float64          // 实际监控的内存使用量
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

// NewMemUsageCollectorDemo 是 MemUsageCollectorDemo 的构造函数，返回一个新的实例
func NewMemUsageCollectorDemo() *MemUsageCollectorDemo {
	return &MemUsageCollectorDemo{
		// 创建一个新的描述符，用于描述内存使用量的指标
		MemUsageDesc: prometheus.NewDesc(
			"memory_usage_bytes",                           // 指标名称
			"系统内存使用量",                                      // 指标帮助信息
			[]string{"instance_id", "instance_name"},       // 动态标签
			prometheus.Labels{"module": "test_demo_usage"}, // 固定标签
		),
		// 设置动态标签的值
		LabelValues: []string{"local", "本地机器"},
	}
}

// updateMemUsage 更新 MemUsageCollectorDemo 实例的内存使用量
func (m *MemUsageCollectorDemo) updateMemUsage() {
	// 使用gopsutil库获取系统内存信息
	memInfo, _ := mem.VirtualMemory()
	// 将可用内存量转换为float64并更新到MemUsage字段
	m.MemUsage = float64(memInfo.Active)
}