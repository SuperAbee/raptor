package servicecenter

import (
	"log"
	"raptor/constants"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

var _ ServiceCenter = (*nacosServiceCenter)(nil)

func newNacosServiceCenter() ServiceCenter {
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
		log.Fatal(err)
	}

	return &nacosServiceCenter{
		nameingClient: namingClient,
	}
}

type nacosServiceCenter struct {
	nameingClient naming_client.INamingClient
}

func (n *nacosServiceCenter) Register(param RegisterParam) (bool, error) {
	return n.nameingClient.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          param.Ip,
		Port:        param.Port,
		ServiceName: param.ServiceName,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		Metadata:    map[string]string{"app": constants.K8S_APP_LABEL},
		GroupName:   constants.NACOS_GROUP,
	})
}

func (n *nacosServiceCenter) GetService(name string) (Service, error) {
	service, err := n.nameingClient.GetService(vo.GetServiceParam{
		ServiceName: name,
		GroupName:   constants.NACOS_GROUP,
	})
	if err != nil {
		return Service{}, err
	}
	var instances []Instance
	for _, i := range service.Hosts {
		instances = append(instances, Instance{
			Ip:      i.Ip,
			Port:    i.Port,
			Healthy: i.Healthy,
		})
	}
	return Service{
		Name:  service.Name,
		Hosts: instances,
	}, nil
}
