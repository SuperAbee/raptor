package configcenter

import (
	"log"
	"raptor/constants"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

var _ ConfigCenter = (*nacosConfigCenter)(nil)

func newNacosConfigCenter() ConfigCenter {
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

	configClient, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	return &nacosConfigCenter{
		configClient: configClient,
	}

}

type nacosConfigCenter struct {
	configClient config_client.IConfigClient
}

func (n *nacosConfigCenter) Save(config Config) (bool, error) {
	return n.configClient.PublishConfig(vo.ConfigParam{
		DataId:  config.ID,
		Group:   constants.NACOS_GROUP,
		Content: config.Content,
	})
}

func (n *nacosConfigCenter) Get(id string) (Config, error) {
	content, err := n.configClient.GetConfig(vo.ConfigParam{
		DataId: id,
		Group:  constants.NACOS_GROUP,
	})

	if err != nil {
		return Config{}, nil
	}

	return Config{
		ID:      id,
		Content: content,
	}, nil
}

func (n *nacosConfigCenter) OnChange(id string, handler func(config Config)) error {
	return n.configClient.ListenConfig(vo.ConfigParam{
		DataId: id,
		Group:  constants.NACOS_GROUP,
		OnChange: func(namespace, group, dataId, data string) {
			handler(Config{
				ID:      dataId,
				Content: data,
			})
		},
	})
}
