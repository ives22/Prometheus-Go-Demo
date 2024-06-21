package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/cpu"
)

type CpuUsageSummaryCollectorDemo struct {
	CpuUsageSummaryDesc *prometheus.Desc
	cpuUsage            uint64
}

// Describe 实现了 Prometheus 的 Collector 接口，用于发送描述符到 Prometheus 的描述符通道
func (c *CpuUsageSummaryCollectorDemo) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.CpuUsageSummaryDesc
}

// Collect 实现了 Prometheus 的 Collector 接口，用于收集指标并发送到 Prometheus 的指标通道
func (c *CpuUsageSummaryCollectorDemo) Collect(ch chan<- prometheus.Metric) {
	// 更新 CPU 使用率
	c.updateCpuUsage()

	// 创建一个新的 Summary 指标并发送到指标通道
	ch <- prometheus.MustNewConstSummary(
		c.CpuUsageSummaryDesc, // 描述符
		1,                     // 采样数
		float64(c.cpuUsage),   // 采样值
		map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001}, // 量化数
	)
}

// NewCpuUsageSummaryCollectorDemo 是 CpuUsageSummaryCollectorDemo 的构造函数，返回一个新的实例
func NewCpuUsageSummaryCollectorDemo() *CpuUsageSummaryCollectorDemo {
	return &CpuUsageSummaryCollectorDemo{
		// 创建一个新的描述符，用于描述 CPU 使用率的汇总指标
		CpuUsageSummaryDesc: prometheus.NewDesc(
			"cpu_usage_summary",
			"系统 CPU 使用率汇总",
			nil, // 无动态标签
			prometheus.Labels{"module": "cpu_usage_summary"}, // 固定标签
		),
	}
}

// updateCpuUsage 更新 CpuUsageSummaryCollectorDemo 实例的 CPU 使用率
func (c *CpuUsageSummaryCollectorDemo) updateCpuUsage() {
	// 获取 CPU 使用率，设置 interval 为 0 表示立即获取当前的 CPU 使用率
	percentages, err := cpu.Percent(0, false)
	if err == nil && len(percentages) > 0 {
		// 更新 cpuUsage 字段为最新的 CPU 使用率（取第一个元素，因为我们只获取总体 CPU 使用率）
		c.cpuUsage = uint64(percentages[0])
	}
}