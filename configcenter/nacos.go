package configcenter

import (
	"github.com/tidwall/gjson"
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
		cache: newMemoryConfigCenter(),
	}

}

type nacosConfigCenter struct {
	configClient config_client.IConfigClient
	cache ConfigCenter
}

func (n *nacosConfigCenter) GetByGroup(id, group string) (Config, error) {
	if r, err := n.cache.GetByGroup(id, group); err == nil && r.ID == id {
		return r, nil
	}
	if group == "" {
		group = constants.NACOS_GROUP
	}

	content, err := n.configClient.GetConfig(vo.ConfigParam{
		DataId: id,
		Group:  group,
	})

	if err != nil {
		return Config{}, nil
	}

	return Config{
		ID:      id,
		Group: group,
		Content: content,
	}, nil
}

func (n *nacosConfigCenter) GetByKV(kv map[string]string, group string) ([]Config, error) {
	if r, err := n.cache.GetByKV(kv, group); err == nil && len(r) != 0 {
		return r, nil
	}
	var ret []Config
	if group == "" {
		group = constants.NACOS_GROUP
	}
	t, err := n.configClient.SearchConfig(vo.SearchConfigParam{
		Search:   "blur",
		Group:    group,
		PageNo:   1,
		PageSize: 1,
	})
	if err != nil {
		return nil, err
	}
	pageIndex := 1
	pageSize := 20
	for i := 0; i < t.TotalCount; {
		c, err := n.configClient.SearchConfig(vo.SearchConfigParam{
			Search:   "blur",
			Group:    group,
			PageNo:   pageIndex,
			PageSize: pageSize,
		})
		if err != nil {
			return nil, err
		}
		pageIndex++
		i += 20

		for _, cc := range c.PageItems {
			match := true
			for k, v := range kv {
				s := cc.Content
				if gjson.Get(s, k).String() != v {
					match = false
					break
				}
			}
			if match {
				ret = append(ret, Config{
					ID:      cc.DataId,
					Content: cc.Content,
				})
			}
		}
	}

	return ret, nil
}

func (n *nacosConfigCenter) Save(config Config) (bool, error) {
	_, err := n.cache.Save(config)
	if err != nil {
		log.Println(err)
	}
	if config.Group == "" {
		config.Group = constants.NACOS_GROUP
	}
	return n.configClient.PublishConfig(vo.ConfigParam{
		DataId:  config.ID,
		Group:   config.Group,
		Content: config.Content,
	})
}

func (n *nacosConfigCenter) Get(id string) (Config, error) {
	if r, err := n.cache.Get(id); err == nil && r.ID == id {
		return r, nil
	}
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
