package monitor

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"testing"
	"time"
)

func TestMonitor(t *testing.T) {
	mockData()

	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(":8877", nil)
	if err != nil {
		log.Println(err)
	}
}

func mockData() {
	go func() {
		for i := 0; ; i++ {
			if i % 2 == 0 {
				TaskProcessedDurations.With(prometheus.Labels{"type": "GET", "target": "x.x.x.x"}).Add(float64(i % 100))
			} else {
				TaskProcessedDurations.With(prometheus.Labels{"type": "GET", "target": "x.x.x.x"}).Add(float64(i % 100))
			}
			time.Sleep(1 * time.Second)
		}
	}()

	go func() {
		for i := 0; ; i++ {
			if i % 2 == 0 {
				TaskProcessing.With(prometheus.Labels{"type": "GET", "target": "x.x.x.x"}).Inc()
			} else {
				TaskProcessing.With(prometheus.Labels{"type": "GET", "target": "x.x.x.x"}).Inc()
			}
			time.Sleep(1 * time.Second)
			if i % 2 == 0 {
				TaskProcessing.With(prometheus.Labels{"type": "GET", "target": "x.x.x.x"}).Dec()
			} else {
				TaskProcessing.With(prometheus.Labels{"type": "GET", "target": "x.x.x.x"}).Dec()
			}
		}
	}()

	go func() {
		for i := 0; ; i++ {
			TaskProcessDurationsHistogram.With(prometheus.Labels{"type": "GET", "target": "x.x.x.x"}).Observe(float64(i % 100))
			time.Sleep(1 * time.Second)
		}
	}()

	go func() {
		for i := 0; ; i++ {
			TaskProcessDurationsSummary.With(prometheus.Labels{"type": "GET", "target": "x.x.x.x"}).Observe(float64(i % 100))
			time.Sleep(1 * time.Second)
		}
	}()
}