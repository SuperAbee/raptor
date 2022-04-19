package eventcenter

import (
	"os"
	"time"
)

func init() {
	if t := os.Getenv("MOCK_EVENT"); t == "true" {
		go func() {
			ec := New()
			ec.Subscribe("Mock", func(event *Event) {
				time.Sleep(20 * time.Second)
			})
			for {
				ec.Publish("Mock", NewEvent())
				time.Sleep(time.Second)
			}
		}()
	}
}
