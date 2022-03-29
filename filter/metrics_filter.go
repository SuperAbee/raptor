package filter

import (
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"raptor/monitor"
	"raptor/proto"
	"time"
)

var (
	_ Filter = (*MetricsPreFilter)(nil)
	_ Filter = (*MetricsPostFilter)(nil)

	MetricsPreFilterKey = "metrics_pre_filter"
	MetricsPostFilterKey = "metrics_post_filter"
)


func init() {
	Filters[MetricsPreFilterKey] = &MetricsPreFilter{}
	Filters[MetricsPostFilterKey] = &MetricsPostFilter{}
}

type MetricsPreFilter struct {}

func (m *MetricsPreFilter) Filter(instance *proto.JobInstance) error {
	log.Println("metrics pre filter begin")
	if instance.Extra == nil {
		instance.Extra = make(map[string]interface{})
	}
	instance.Extra["start_time"] = time.Now()

	monitor.TaskProcessing.With(prometheus.Labels{
		"type": instance.Config.Task.Type,
		"target": instance.Config.TargetService,
	}).Inc()

	return nil
}

type MetricsPostFilter struct {}

func (m *MetricsPostFilter) Filter(instance *proto.JobInstance) error {
	log.Println("metrics post filter begin")
	start := instance.Extra["start_time"].(time.Time)
	duration := ((float64)(time.Now().UnixNano() - start.UnixNano())) / 1000000000

	labels := prometheus.Labels{
		"type": instance.Config.Task.Type,
		"target": instance.Config.TargetService,
	}
	monitor.TaskProcessing.With(labels).Dec()
	monitor.TaskProcessedDurations.With(labels).Add(duration)
	monitor.TaskProcessDurationsHistogram.With(labels).Observe(duration)
	monitor.TaskProcessDurationsSummary.With(labels).Observe(duration)

	return nil
}



