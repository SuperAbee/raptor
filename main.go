package main

import (
	"log"
	"raptor/configcenter"
	"raptor/servicecenter"
	"time"
)

func main() {
	sc, _ := servicecenter.New(servicecenter.Nacos)

	sc.Register(servicecenter.RegisterParam{
		Ip:          "10.0.0.2",
		Port:        1234,
		ServiceName: "demo",
	})

	sc.Register(servicecenter.RegisterParam{
		Ip:          "10.0.0.3",
		Port:        1234,
		ServiceName: "demo",
	})

	service, _ := sc.GetService("demo")
	log.Println(service)

	cc, _ := configcenter.New(configcenter.Nacos)

	cc.Save(configcenter.Config{
		ID:      "config-1",
		Content: "hello time " + time.Now().String(),
	})
	time.Sleep(time.Second)

	config, _ := cc.Get("config-1")
	log.Printf("%+v\n", config)
	time.Sleep(time.Second)

	cc.OnChange("config-1", func(config configcenter.Config) {
		log.Printf("config update: %s\n", config)
	})

	cc.Save(configcenter.Config{
		ID:      "config-1",
		Content: "hello time " + time.Now().String(),
	})
	time.Sleep(time.Second)
}
