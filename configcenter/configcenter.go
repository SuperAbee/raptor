package configcenter

import "fmt"

type ConfigCenterType string

const (
	K8S   = "k8s"
	Nacos = "nacos"
)

func New(ccType ConfigCenterType) (ConfigCenter, error) {
	switch ccType {
	case K8S:
		return newK8sConfigCenter(), nil
	case Nacos:
		return newNacosConfigCenter(), nil
	}
	return nil, fmt.Errorf("ConfigCenterType '%v' not supported", ccType)
}

type ConfigCenter interface {
	Get(name string) Config
}

type Config struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
