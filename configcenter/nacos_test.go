package configcenter

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"testing"
	"time"
)

func TestNacosConfigCenter(t *testing.T) {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		err := http.ListenAndServe(":8877", nil)
		if err != nil {
			log.Println(err)
		}
	}()
	time.Sleep(time.Second)
	cc := newNacosConfigCenter()
	cc.Save(Config{
		ID:      "1",
		Group:   "",
		Content: "11111111111111111111111111111111111111111",
	})
	cc.Save(Config{
		ID:      "2",
		Group:   "",
		Content: "111111111111111111111111111111111111111111" +
			"111111111111111111111111111111111111111111111111111" +
			"11111111111111111111111111111111111111111111111111111111" +
			"1111111111111111111111111111111111111111111111111111111111111" +
			"111111111111111111111111111111111111111111111111111111111111111111" +
			"11111111111111111111111111111111111111111111111",
	})

	for i := 0; ; i++ {
		if i % 2 == 0 {
			cc.Get("1")
		} else {
			cc.Get("2")
		}

		time.Sleep(time.Second)
	}
}
