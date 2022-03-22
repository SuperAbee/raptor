package configcenter

import (
	"log"
	"os"
	"sync"
)

var (
	Type string
	cc ConfigCenter
	once = sync.Once{}
)

const (
	K8S   = "k8s"
	Nacos = "nacos"
	Memory = "memory"
)

func New() ConfigCenter {
	once.Do(func() {
		if t := os.Getenv("CONFIG_CENTER"); t != "" {
			log.Printf("CONFIG_CENTER: %v\n", t)
			Type = t
		}
		switch Type {
		case K8S:
			cc = newK8sConfigCenter()
		case Nacos:
			cc = newNacosConfigCenter()
		default:
			cc = newMemoryConfigCenter()
		}
	})
	return cc
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
