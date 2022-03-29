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
)
