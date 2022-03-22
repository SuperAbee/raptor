package servicecenter

import (
	"log"
	"os"
	"sync"
)

const (
	K8S   = "k8s"
	Nacos = "nacos"
	Memory = "memory"
)

var (
	Type string
	sc ServiceCenter
	once = sync.Once{}
)

func New() ServiceCenter {
	once.Do(func() {
		if t := os.Getenv("SERVICE_CENTER"); t != "" {
			log.Printf("SERVICE_CENTER: %v\n", t)
			Type = t
		}
		switch Type {
		case K8S:
			sc = newK8sServiceCenter()
		case Nacos:
			sc = newNacosServiceCenter()
		default:
			sc = newMemoryServiceCenter()
		}
	})
	return sc

}

type ServiceCenter interface {
	Register(param RegisterParam) (bool, error)
	GetService(name string) (Service, error)
}

type RegisterParam struct {
	Ip          string
	Port        uint64
	ServiceName string
}

type Service struct {
	Name  string     `json:"name"`
	Hosts []Instance `json:"hosts"`
}

type Instance struct {
	Ip      string `json:"ip"`
	Port    uint64 `json:"port"`
	Healthy bool   `json:"healthy"`
}
