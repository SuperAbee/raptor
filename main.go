package main

import (
	"raptor/router"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.New()
	router.Route(r)

	r.Run(":1234")

	// sc.Register(servicecenter.RegisterParam{
	// 	Ip:          "127.0.0.1",
	// 	Port:        8888,
	// 	ServiceName: "demo",
	// })

	// service, _ := sc.GetService("demo")
	// log.Printf("%+v\n", service)

	// cc, _ := configcenter.New(configcenter.Nacos)

	// cc.Save(configcenter.Config{
	// 	ID:      "config-1",
	// 	Content: "hello time " + time.Now().String(),
	// })
	// time.Sleep(time.Second)

	// config, _ := cc.Get("config-1")
	// log.Printf("%+v\n", config)
	// time.Sleep(time.Second)

	// cc.OnChange("config-1", func(config configcenter.Config) {
	// 	log.Printf("config update: %+v\n", config)
	// })

	// cc.Save(configcenter.Config{
	// 	ID:      "config-1",
	// 	Content: "hello time " + time.Now().String(),
	// })
	// time.Sleep(time.Second * 1000)

}
