package main

import (
	"raptor/servicecenter"
	"time"
)

func main() {
	// servicecenter, _ := servicecenter.NewServiceCenter(servicecenter.Nacos)
	// servicecenter.GetService("demo")

	for {
		servicecenter.K8SNamingTest()
		time.Sleep(10 * time.Second)
	}
}
