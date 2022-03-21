package eventcenter

import "sync"

const TOPIC = "topic"

func New() *EventCenter {
	r := &EventCenter{
		eventCh:     make(chan *Event, 100),
		subscribers: make(map[string][]Handler),
		mu:          sync.RWMutex{},
	}
	r.start()
	return r
}

type EventCenter struct {
	eventCh     chan *Event
	subscribers map[string][]Handler
	mu          sync.RWMutex
}

type Handler func(event *Event)

func (e *EventCenter) Subscribe(topic string, handler Handler) {
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
	e.eventCh <- event.WithHeader(TOPIC, topic)
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
				for _, h := range t {
					h(event)
				}
			}
		}
	}()
}
