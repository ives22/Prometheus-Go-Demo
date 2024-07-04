<div style="text-align: center; font-size: 28px; color: #2096FF">Prometheus Exporter开发</div>

# 说明

在日常监控工作中，虽然prometheus和第三方已经给我们提供了很多使用的exporter，比如mysql有mysql_exporter、redis有redis_exporter等等，但有时候由于我们项目的特定场景及特定的需求，我们需要自定义来精细化实现监控指标。确保得到想要的结果。

`Prometheus`的`Server`端, 只认如下数据格式:

```yaml
# HELP go_goroutines Number of goroutines that currently exist.
# TYPE go_goroutines gauge
go_goroutines 19
```

但是Prometheus客户端本身也提供一些简单数据二次加工的能力, 他把这种能力描述为4种指标类型: [详情参考](../../../Monitor/Prometheus/chapter1.md)

+ Gauges（仪表盘）：Gauge类型代表一种样本数据可以任意变化的指标，即可增可减。
+ Counters（计数器）：counter类型代表一种样本数据单调递增的指标，即只增不减，除非监控系统发生了重置。
+ Histograms（直方图）：创建直方图指标比 counter 和 gauge 都要复杂，因为需要配置把观测值归入的 bucket 的数量，以及每个 bucket 的上边界。Prometheus 中的直方图是累积的，所以每一个后续的 bucket 都包含前一个 bucket 的观察计数，所有 bucket 的下限都从 0 开始的，所以我们不需要明确配置每个 bucket 的下限，只需要配置上限即可。
+ Summaries（摘要）：与Histogram类似类型，用于表示一段时间内的数据采样结果（通常是请求持续时间或响应大小等），但它直接存储了分位数（通过客户端计算，然后展示出来），而不是通过区间计算

# 指标采集

## Gauges

着是最常见的`Metric`类型, 也就是我们说的实时指标, 值是什么就返回什么, 并不会进行加工处理

SDK提供了该指标的构造函数: `NewGauge`

```go
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
```

`Gauge`对象提供了如下方法用来设置它的值:

```go
// 使用 Set() 设置指定的值
queueLength.Set(0)

queueLength.Inc()   // +1: gauge增加1
queueLength.Dec()   // -1: gauge减少1
queueLength.Add(20) // 增加20个增量
queueLength.Sub(12) // 减少12个
```

**完整示例**

```go
package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

func main() {
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

	// 注册指标
	prometheus.MustRegister(queueLength)

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":8050", nil)
}
```

运行上面代码后，访问 http://localhost:8050/metrics ，会得到如下结果

```yaml
...
# HELP system_web_queue_length The number of items in the queue
# TYPE system_web_queue_length gauge
system_web_queue_length{module="http-server"} 8
```

## Counter

Counter是计算器指标, 用于统计次数使用, 通过 prometheus.NewCounter() 函数来初始化指标对象。

```go
totalRequests := prometheus.NewCounter(prometheus.CounterOpts{
  Name:      "http_requests_total",
  Help: "The total number of handled HTTP requests.",
})
```

Counter只是提供如下两个方法

```go
totalRequests.Inc()   // 计数器增加1
totalRequests.Add(2)  // 计数器增加n
```

**完整示例：**

```go
package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

func main() {
	totalRequests := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "The total number of handled HTTP requests.",
	})

	for i := 0; i <= 10; i++ {
		totalRequests.Inc()
	}

	// 注册指标
	prometheus.MustRegister(totalRequests)

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":8050", nil)
}
```

运行上面代码后，访问 http://localhost:8050/metrics ，会得到如下结果

```yaml
...
# HELP http_requests_total The total number of handled HTTP requests.
# TYPE http_requests_total counter
http_requests_total 11
```

## Histogram

Histograms 被叫主直方图或者柱状图, 主要用于统计指标值的一个分布情况, 也就是常见的概率统计问题

比如, 我们要统计一个班级的 成绩分布情况：

![image-20240704210821356](assets/image-20240704210821356.png)

+ 横轴表示 分数的区间(0~50, 50~55, ...)
+ 纵轴表示 落在该区间的人数

prometheus的Histograms也是用于解决这类问题的, 在prometheus里面用于设置横轴区间的概念叫Bucket, 不同于传统的区间设置之处, 在于prometheus的Bucket只能设置上限, 下线就是最小值，也就是说 换用prometheus Histograms, 我们上面的区间会变成这样:

```
0 ~ 50
0 ~ 55
0 ~ 60
...
```

可以看出当我们设置好了Bucket后, prometheus的客户端需要统计落入每个Bucket中的值得数量(也就是一个Counter), 也就是Histograms这种指标类型的计算逻辑。

在监控里面, Histograms 典型的应用场景 就是统计请求耗时分布, 比如

```
0 ~ 100ms 请求个数
0 ~ 500ms 请求个数
0 ~ 5000ms 请求个数
```

使用`NewHistogram`初始化一个直方图类型的指标:

```go
requestDurations := prometheus.NewHistogram(prometheus.HistogramOpts{
  Name: "http_request_duration_seconds",
  Help: "A histogram of the HTTP request durations in seconds.",
  // Bucket 配置：第一个 bucket 包括所有在 0.05s 内完成的请求，最后一个包括所有在10s内完成的请求。
  Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
})
```

Histogram类型指标提供一个 `Observe()` 方法, 用于加入一个值到直方图中, 当然加入后体现在直方图中的不是具体的值，而是值落入区间的统计，实际上每个bucket 就是一个 Counter指标

**完整示例：**

```go
package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

func main() {
	requestDurations := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "http_request_duration_seconds",
		Help: "A histogram of the HTTP request durations in seconds.",
		// Bucket 配置：第一个 bucket 包括所有在 0.05s 内完成的请求，最后一个包括所有在10s内完成的请求。
		Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	})

	// 添加值
	for _, v := range []float64{0.01, 0.02, 0.3, 0.4, 0.6, 0.7, 5.5, 11} {
		requestDurations.Observe(v)
	}

	// 注册指标
	prometheus.MustRegister(requestDurations)

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":8050", nil)
}
```

运行上面代码后，访问 http://localhost:8050/metrics ，会得到如下结果

```yaml
...
# HELP http_request_duration_seconds A histogram of the HTTP request durations in seconds.
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{le="0.05"} 2
http_request_duration_seconds_bucket{le="0.1"} 2
http_request_duration_seconds_bucket{le="0.25"} 2
http_request_duration_seconds_bucket{le="0.5"} 4
http_request_duration_seconds_bucket{le="1"} 6
http_request_duration_seconds_bucket{le="2.5"} 6
http_request_duration_seconds_bucket{le="5"} 6
http_request_duration_seconds_bucket{le="10"} 7
http_request_duration_seconds_bucket{le="+Inf"} 8
http_request_duration_seconds_sum 18.53
http_request_duration_seconds_count 8
```

注意点:

+ le="+Inf", 表示小于正无穷, 也就是统计所有的含义
+ 后缀 _sum,  参加统计的值的求和
+ 后缀 _count 参加统计的值得总数

很多时候直接依赖直方图还是很难定位问题, 因为很多时候，我们需要的是请求的一个概统分布, 比如百分之99的请求 落在了那个区间(比如99%请求都在500ms内完成的), 从而判断我们的访问 从整体上看 是良好的。

而像上面的概念分布问题有一个专门的名称叫: quantile, 翻译过来就分位数, 及百分之多少的请求 在那个范围下

那基于直方图提供的数据, 我们是可以计算出分位数的, 但是这个分位数的精度 会受到分区设置精度的影响(bucket设置)， 比如你如果只设置了2个bucket, 0.001, 5, 那么你统计出来的100%这个分位数 就是5s, 因为所有的请求都会落到这个bucket中

如果我们的bucket设置是合理的, 我又想使用直方图来统计分位数喃? prometheus的QL, 提供了专门的函数histogram_quantile, 可以用于基于直方图的统计数据，计算分位数。

如果服务端压力很大, bucket也不确定, 我能不能直接在客户端计算分位数(quantile)喃?

答案是有的，就是第四种指标类型: Summaries

## Summary

这种类型的指标 就是用于计算分位数(quantile)的, 因此他需要配置一个核心参数: 你需要统计那个(百)分位

用NewSummary来构建该类指标

```go
requestDurations := prometheus.NewSummary(prometheus.SummaryOpts{
    Name:       "http_request_duration_seconds",
    Help:       "A summary of the HTTP request durations in seconds.",
    Objectives: map[float64]float64{
      0.5: 0.05,   // 第50个百分位数，最大绝对误差为0.05。
      0.9: 0.01,   // 第90个百分位数，最大绝对误差为0.01。
      0.99: 0.001, // 第90个百分位数，最大绝对误差为0.001。
    },
  },
)
```

和直方图一样, 他也近提供一个方法: `Observe`, 用于统计数据

**完整示例：**

```go
package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

func main() {
	requestDurations := prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "http_request_duration_seconds",
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

	// 注册指标
	prometheus.MustRegister(requestDurations)

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":8050", nil)
}
```

运行上面代码后，访问 http://localhost:8050/metrics ，会得到如下结果

```yaml
...
# HELP http_request_duration_seconds A summary of the HTTP request durations in seconds.
# TYPE http_request_duration_seconds summary
http_request_duration_seconds{quantile="0.5"} 0.4
http_request_duration_seconds{quantile="0.9"} 11
http_request_duration_seconds{quantile="0.99"} 11
http_request_duration_seconds_sum 18.53
http_request_duration_seconds_count 8
```

可以看出来 直接使用客户端计算分位数, 准确度不依赖我们设置bucket, 是比较推荐的做法

# 指标标签

`Prometheus`将指标的标签分为2类:

+ 静态标签: `constLabels`, 在指标创建时, 就提前声明好, 采集过程中永不变动
+ 动态标签: `variableLabels`, 用于在指标的收集过程中动态补充标签, 比如kafka集群的exporter需要动态补充 instance_id

静态标签在NewGauge之类时已经指明, 下面讨论下如何添加动态标签。

要让你的指标支持动态标签有专门的构造函数, 对应关系如下:

+ NewGauge() 变成 `NewGaugeVec()`
+ NewCounter() 变成 `NewCounterVec()`
+ NewSummary() 变成 `NewSummaryVec()`
+ NewHistogram() 变成 `NewHistogramVec()`

下面以`NewGaugeVec`为例进行示例

NewGaugeVec相比于NewGauge只多出了一个**labelNames**的参数:

```go
func NewGaugeVec(opts GaugeOpts, labelNames []string) *GaugeVec
```

一定声明了labelNames, 我们在为指标设置值得时候就必须带上对应个数的标签(一一对应, 二维数组)

```go
queueLength.WithLabelValues("rm_001", "kafka01").Set(100)
```

**完整示例：**

```go
package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

func main() {
	queueLength := prometheus.NewGaugeVec(prometheus.GaugeOpts{
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
	}, []string{"instance_id", "instance_name"})

	queueLength.WithLabelValues("rm_001", "kafka01").Set(100)

	// 注册指标
	prometheus.MustRegister(queueLength)

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":8050", nil)
}
```

运行上面代码后，访问 http://localhost:8050/metrics ，会得到如下结果

```yaml
# HELP system_web_queue_length The number of items in the queue
# TYPE system_web_queue_length gauge
system_web_queue_length{instance_id="rm_001",instance_name="kafka01",module="http-server"} 100
```



# 指标注册

我们把指标采集完成后需要注册给Prometheus的`Http Handler`这样才能暴露出去, 好在Prometheus的客户端给我们提供了对于的接口

`Prometheus` 定义了一个注册表的接口:

```go
// 指标注册接口
type Registerer interface {
	// 注册采集器, 有异常会报错
	Register(Collector) error
	// 注册采集器, 有异常会panic
	MustRegister(...Collector)
	// 注销该采集器
	Unregister(Collector) bool
}
```

## 默认注册表

`Prometheus` 实现了一个默认的`Registerer`对象, 也就是默认注册表

```go
var (
	defaultRegistry              = NewRegistry()
	DefaultRegisterer Registerer = defaultRegistry
	DefaultGatherer   Gatherer   = defaultRegistry
)
```

我们通过`prometheus`提供的`MustRegister`可以将我们自定义指标注册进去

```go
// 在默认的注册表中注册该指标
prometheus.MustRegister(temp)
prometheus.Register()
prometheus.Unregister()
```

下面时一个完整的例子

```go
package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

func main() {
	// 创建一个 gauge 类型的指标
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

	// 在默认的注册表中注册该指标
	prometheus.MustRegister(queueLength)

	// 设置 gauge 的值为 100
	queueLength.Set(100)

	// 暴露指标
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":8050", nil)
}
```

启动后重新访问指标接口 http://localhost:8050/metrics，仔细对比我们会发现多了一个名为 system_web_queue_length 的指标:

```yaml
...
# HELP system_web_queue_length The number of items in the queue
# TYPE system_web_queue_length gauge
system_web_queue_length{instance_id="rm_001",instance_name="kafka01",module="http-server"} 100
...
```

## 自定义注册表

`Prometheus` 默认的`Registerer`, 会添加一些默认指标的采集, 比如看到的`go`运行时和当前`process`相关信息, 如果不想采集指标, 那么最好的方式是使用自定义的注册表

```go
package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

func main() {
	// 创建一个 gauge 类型的指标
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

	// 创建一个自定义的注册表
	registry := prometheus.NewRegistry()
	// 在自定义的的注册表中注册该指标
	registry.MustRegister(queueLength)

	// 设置 gauge 的值为 100
	queueLength.Set(100)

	// 暴露指标
	http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{Registry: registry}))
	http.ListenAndServe(":8050", nil)
}
```

+ 使用`prometheus.NewRegistry()`创建一个全新的注册表
+ 通过注册表对象的`MustRegister`把指标注册到自定义的注册表中

**暴露指标的时候必须通过**调用 `promhttp.HandleFor()` 函数来创建一个专门针对我们自定义注册表的 `HTTP` 处理器，我们还需要在 `promhttp.HandlerOpts` 配置对象的 `Registry` 字段中传递我们的注册表对象

```go
...
// 暴露指标
http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{Registry: registry}))
http.ListenAndServe(":8050", nil)
```

最后我们看到我们的指标少了很多, 除了`promhttp_metric_handler`就只有我们自定义的指标了

```
# HELP promhttp_metric_handler_errors_total Total number of internal errors encountered by the promhttp metric handler.
# TYPE promhttp_metric_handler_errors_total counter
promhttp_metric_handler_errors_total{cause="encoding"} 0
promhttp_metric_handler_errors_total{cause="gathering"} 0
# HELP system_web_queue_length The number of items in the queue
# TYPE system_web_queue_length gauge
system_web_queue_length{module="http-server"} 100
```

那如果后面又想把go运行时和当前process相关加入到注册表中暴露出去怎么办?

其实Prometheus在客户端中默认有如下Collector供我们选择

![](assets/prom_collector.png)

只需把我们需要的添加到我们自定义的注册表中即可

```go
 // 添加 process 和 Go 运行时指标到我们自定义的注册表中
 registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
 registry.MustRegister(prometheus.NewGoCollector())
```

然后我们再次访问http://localhost:8050/metrics, 是不是发现之前少的指标又回来了

通过查看prometheus提供的Collectors我们发现, 直接把指标注册到registry中的方式不太优雅, 为了能更好的模块化, 我们需要把指标采集封装为一个Collector对象, 这也是很多第三方Collecotor的标准写法。



## demo采集器

实现`demo`采集器

```go
type DemoCollector struct {
	queueLengthDesc *prometheus.Desc
	labelValues     []string
}

func NewDemoCollector() *DemoCollector {
	return &DemoCollector{
		queueLengthDesc: prometheus.NewDesc(
			"system_web_demo_queue_length",
			"The number of items in the queue.",
			// 动态标签的key列表
			[]string{"instnace_id", "instnace_name"},
			// 静态标签
			prometheus.Labels{"module": "http-server"},
		),
		// 动态标的value列表, 这里必须与声明的动态标签的key一一对应
		labelValues: []string{"mq_001", "kafka01"},
	}
}

func (c *DemoCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.queueLengthDesc
}

func (c *DemoCollector) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(c.queueLengthDesc, prometheus.GaugeValue, 100, c.labelValues...)
}
```

重构后我们的代码将变得简洁优雅:

```go
package main

import (
 "net/http"

 "github.com/prometheus/client_golang/prometheus"
 "github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// 创建一个自定义的注册表
	registry := prometheus.NewRegistry()

	// 注册自定义采集器
	registry.MustRegister(NewDemoCollector())

	// 暴露指标
	http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{Registry: registry}))
	http.ListenAndServe(":8050", nil)
 }
```

最后我们看到的结果如下:

```yaml
# HELP promhttp_metric_handler_errors_total Total number of internal errors encountered by the promhttp metric handler.
# TYPE promhttp_metric_handler_errors_total counter
promhttp_metric_handler_errors_total{cause="encoding"} 0
promhttp_metric_handler_errors_total{cause="gathering"} 0
# HELP system_web_demo_queue_length The number of items in the queue.
# TYPE system_web_demo_queue_length gauge
system_web_demo_queue_length{instnace_id="mq_001",instnace_name="kafka01",module="http-server"} 100
```