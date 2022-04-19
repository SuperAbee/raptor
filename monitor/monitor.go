package monitor

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	TaskProcessedDurations = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "task_processed_seconds",
		Help: "任务累计执行的时间总和",
	}, []string{"type", "target"})

	TaskProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "task_processed_total",
		Help: "任务累计执行的次数",
	}, []string{"type", "target"})

	TaskProcessing = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "task_processing_total",
		Help: "当前正在执行的任务总数",
	}, []string{"type", "target"})

	TaskProcessDurationsHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:        "task_duration_histogram_seconds",
		Help:        "任务执行耗时的柱状图",
		Buckets:     []float64{10, 20, 50, 100},
	}, []string{"type", "target"})

	TaskProcessDurationsSummary = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name:        "task_duration_summary_seconds",
		Help:        "任务执行耗时的分位图",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	}, []string{"type", "target"})

	TaskDelaySummary = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name:        "task_delay_summary_seconds",
		Help:        "任务延迟执行时间的分位图",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	}, []string{"type", "target"})

	EventSubscribers = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "event_subscribers",
		Help: "事件订阅者数量",
	}, []string{"topic"})

	EventGoroutineUsing = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "event_goroutine_using_total",
		Help: "事件推送协程数",
	})

	EventPublished = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "event_published_total",
		Help: "事件累计推送总数",
	}, []string{"topic"})

	EventConsumed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "event_consumed_total",
		Help: "事件累计消费总数",
	}, []string{"topic"})

	EventConsumeDurationsHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:        "event_duration_histogram_nanoseconds",
		Help:        "事件推送耗时的柱状图",
		Buckets:     []float64{100, 200, 500, 1000},
	}, []string{"topic"})

	EventConsumedDurationsSummary = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name:        "event_duration_summary_nanoseconds",
		Help:        "事件推送耗时的分位图",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	}, []string{"topic"})

	EventDelayDurationsHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:        "event_delay_duration_histogram_nanoseconds",
		Help:        "事件延迟推送耗时的柱状图",
		Buckets:     []float64{1000, 2000, 5000, 10000},
	}, []string{"topic"})

	EventDelayDurationsSummary = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name:        "event_delay_duration_summary_nanoseconds",
		Help:        "事件延迟推送耗时的分位图",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	}, []string{"topic"})

	ConfigCenterQuery = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "config_center_query_total",
		Help: "配置中心查询次数",
	}, []string{"group"})

	ConfigCenterCacheHit = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "config_center_cache_hit_total",
		Help: "配置中心缓存命中次数",
	}, []string{"group"})

	ConfigCenterContentLengthHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:        "config_center_content_length_histogram",
		Help:        "配置中心查询结果大小的柱状图",
		Buckets:     []float64{50, 100, 200, 500},
	}, []string{"group"})

	ConfigCenterContentLengthSummary = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name:        "config_center_content_length_summary",
		Help:        "配置中心查询结果大小的分位图",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	}, []string{"group"})

	ConfigCenterDurationHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:        "config_center_duration_histogram_nanosecond",
		Help:        "配置中心查询时间的柱状图",
		Buckets:     []float64{500, 1000, 2000, 5000},
	}, []string{"group"})

	ConfigCenterDurationSummary = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name:        "config_center_duration_summary_nanosecond",
		Help:        "配置中心查询时间的分位图",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	}, []string{"group"})
)
