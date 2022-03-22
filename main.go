package main

import (
	"flag"
	"fmt"
	"os"
	"raptor/configcenter"
	"raptor/router"
	"raptor/servicecenter"

	"github.com/gin-gonic/gin"
)

var (
	sc = flag.String("servicecenter", servicecenter.Memory, "Type of Service Center, Memory for Default.")
	cc = flag.String("configcenter", configcenter.Memory, "Type of Config Center, Memory for Default.")
)

func main() {
	flag.Parse()
	servicecenter.Type = *sc
	configcenter.Type = *cc

	fmt.Println(servicecenter.Type)
	fmt.Println(configcenter.Type)

	for index, arg := range os.Args {
		fmt.Printf("args[%d]=%v\n", index, arg)
	}

	r := gin.New()
	router.Route(r)

	r.Run(":1234")

}
