package servicecenter

import (
	"log"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

var _ ServiceCenter = (*nacosServiceCenter)(nil)

func newNacosServiceCenter() ServiceCenter {
	return &nacosServiceCenter{}
}

type nacosServiceCenter struct {
}

func (n *nacosServiceCenter) GetService(name string) Service {
	return Service{}
}

// ignore
func NacosNamingTest() {
	clientConfig := constant.ClientConfig{
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		RotateTime:          "1h",
		MaxAge:              3,
		LogLevel:            "debug",
	}

	serverConfigs := []constant.ServerConfig{
		{
			IpAddr:      "127.0.0.1",
			ContextPath: "/nacos",
			Port:        8848,
			Scheme:      "http",
		},
	}

	namingClient, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		log.Fatalf("new naming client fail: %v", err)
	}

	success, err := namingClient.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          "10.0.0.11",
		Port:        8848,
		ServiceName: "demo.go",
		Weight:      10,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		Metadata:    map[string]string{"idc": "shanghai"},
	})
	if err != nil || !success {
		log.Fatalf("register fail: %v", err)
	}

	services, err := namingClient.GetAllServicesInfo(vo.GetAllServiceInfoParam{PageSize: 10})
	if err != nil {
		log.Fatalf("get services fail: %v", err)
	}

	for _, serviceName := range services.Doms {
		service, err := namingClient.GetService(vo.GetServiceParam{ServiceName: serviceName})
		if err != nil {
			log.Printf("get service %v fail: %v", serviceName, err)
		}
		log.Println(service)
	}

	// configClient, err := clients.NewConfigClient(
	// 	vo.NacosClientParam{
	// 		ClientConfig:  &clientConfig,
	// 		ServerConfigs: serverConfigs,
	// 	},
	// )
}
