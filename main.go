package main

import (
	"github.com/gin-gonic/gin"
	"raptor/router"
)

func main() {
	r := gin.New()
	router.Route(r)

	r.Run(":1234")
}
