package main

import "raptor/servicecenter"

func main() {
	servicecenter, _ := servicecenter.NewServiceCenter(servicecenter.Nacos)
	servicecenter.GetService("demo")
}
