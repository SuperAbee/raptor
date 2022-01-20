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
	Save(config Config) (bool, error)
	Get(id string) (Config, error)
	OnChange(id string, handler func(config Config)) error
}

type Config struct {
	ID      string `json:"id"`
	Content string `json:"content"`
}
