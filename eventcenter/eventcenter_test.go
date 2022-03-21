package eventcenter

import (
	"fmt"
	"testing"
	"time"
)

func TestEventCenter(t *testing.T) {
	es := New()
	es.Subscribe("t1", func(event *Event) {
		fmt.Printf("s1: %v\n", event)
	})
	es.Subscribe("t1", func(event *Event) {
		fmt.Printf("s2: %v\n", event)
	})

	e := NewEvent().WithHeader("a", "b").WithBody("hello")
	es.Publish("t1", e)

	time.Sleep(time.Second)
}
