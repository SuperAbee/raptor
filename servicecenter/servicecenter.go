package servicecenter

import "fmt"

type ServiceCenterType string

const (
	K8S   = "k8s"
	Nacos = "nacos"
)

func New(scType ServiceCenterType) (ServiceCenter, error) {
	switch scType {
	case K8S:
		return newK8sServiceCenter(), nil
	case Nacos:
		return newNacosServiceCenter(), nil
	}
	return nil, fmt.Errorf("ServiceCenterType '%v' not supported", scType)
}

type ServiceCenter interface {
	Get(name string) Service
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
