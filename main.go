package main

import (
	"log"
	"net/http"
	"raptor/router"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		err := http.ListenAndServe(":8877", nil)
		if err != nil {
			log.Println(err)
		}
	}()

	r := gin.New()
	router.Route(r)

	r.Run(":1234")
}
