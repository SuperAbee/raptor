package configcenter

import (
	"github.com/tidwall/gjson"
	"sync"
)

var (
	_ ConfigCenter = (*memoryConfigCenter)(nil)
	defaultGroup = "DEFAULT"
)

// newMemoryConfigCenter for test
func newMemoryConfigCenter() ConfigCenter {
	return &memoryConfigCenter{
		configMux: sync.RWMutex{},
		configs:   make(map[string]Config),
		changeCh:  make(chan Config, 128),
		handlerMux: sync.RWMutex{},
		handlers: make(map[string]func(config Config)),
	}
}

type memoryConfigCenter struct {
	configMux sync.RWMutex
	configs   map[string]Config
	changeCh  chan Config

	handlerMux sync.RWMutex
	handlers map[string]func(config Config)
}

func (m *memoryConfigCenter) GetByGroup(id, group string) (Config, error) {
	return m.Get(id)
}

func (m *memoryConfigCenter) GetByKV(kv map[string]string, group string) ([]Config, error) {
	if group == "" {
		group = defaultGroup
	}
	m.configMux.RLock()
	m.configMux.RUnlock()
	var ret []Config
	for _, config := range m.configs {
		match := true
		for k, v := range kv {
			if config.Group != group {
				match = false
				break
			}
			if gjson.Get(config.Content, k).String() != v {
				match = false
				break
			}
		}
		if match {
			ret = append(ret, config)
		}
	}
	return ret, nil
}

func (m *memoryConfigCenter) Save(config Config) (bool, error) {
	if config.Group == "" {
		config.Group = defaultGroup
	}
	m.configMux.Lock()
	defer m.configMux.Unlock()
	m.configs[config.ID] = config

	m.handlerMux.RLock()
	defer m.handlerMux.RUnlock()
	if h := m.handlers[config.ID]; h != nil {
		h(config)
	}

	return true, nil
}

func (m *memoryConfigCenter) Get(id string) (Config, error) {
	m.configMux.RLock()
	defer m.configMux.RUnlock()
	return m.configs[id], nil
}

func (m *memoryConfigCenter) OnChange(id string, handler func(config Config)) error {
	m.handlerMux.Lock()
	defer m.handlerMux.Unlock()
	m.handlers[id] = handler
	return nil
}
