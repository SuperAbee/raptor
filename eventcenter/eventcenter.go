package eventcenter

import (
	"github.com/prometheus/client_golang/prometheus"
	"raptor/monitor"
	"sync"
	"time"
)

const (
	TOPIC = "topic"
	DELAY = "delay"
)

var once sync.Once
var eventCenter *EventCenter

func New() *EventCenter {
	once.Do(func() {
		eventCenter = &EventCenter{
			eventCh:     make(chan *Event, 100),
			subscribers: make(map[string][]Handler),
			mu:          sync.RWMutex{},
			ch:			 make(chan struct{}, 20),
		}
		eventCenter.start()
	})

	return eventCenter
}

type EventCenter struct {
	eventCh     chan *Event
	subscribers map[string][]Handler
	mu          sync.RWMutex
	ch 			chan struct{}
}

type Handler func(event *Event)

func (e *EventCenter) Subscribe(topic string, handler Handler) {
	monitor.EventSubscribers.With(prometheus.Labels{"topic": topic}).Inc()
	e.mu.Lock()
	defer e.mu.Unlock()

	s := e.subscribers[topic]
	s = append(s, handler)
	e.subscribers[topic] = s
}

func (e *EventCenter) Publish(topic string, event *Event) {
	if event == nil {
		return
	}
	monitor.EventPublished.With(prometheus.Labels{"topic": topic}).Inc()
	e.eventCh <- event.WithHeader(TOPIC, topic).
		WithHeader(DELAY, time.Now().Format(time.RFC3339Nano))
}

func (e *EventCenter) start() {
	go func() {
		for {
			select {
			case event := <-e.eventCh:
				e.mu.RLock()
				s := e.subscribers[event.Header[TOPIC]]
				t := make([]Handler, len(s))
				copy(t, s)
				e.mu.RUnlock()
				e.ch<- struct{}{}
				monitor.EventGoroutineUsing.Inc()
				go func() {
					start := time.Now()
					for _, h := range t {
						h(event)
					}
					<-e.ch
					monitor.EventGoroutineUsing.Dec()
					monitor.EventConsumed.With(prometheus.Labels{"topic": event.Header[TOPIC]}).Inc()

					end := time.Now()
					delay, err := time.Parse(time.RFC3339Nano, event.Header[DELAY])
					if err != nil {
						panic(err)
					}
					monitor.EventConsumeDurationsHistogram.
						With(prometheus.Labels{"topic": event.Header[TOPIC]}).
						Observe(float64(end.Sub(start)))
					monitor.EventConsumedDurationsSummary.
						With(prometheus.Labels{"topic": event.Header[TOPIC]}).
						Observe(float64(end.Sub(start)))
					monitor.EventDelayDurationsHistogram.
						With(prometheus.Labels{"topic": event.Header[TOPIC]}).
						Observe(float64(end.Sub(delay)))
					monitor.EventDelayDurationsSummary.
						With(prometheus.Labels{"topic": event.Header[TOPIC]}).
						Observe(float64(end.Sub(delay)))
				}()
			}
		}
	}()
}
