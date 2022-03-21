package configcenter

import (
	"sync"
)

var _ ConfigCenter = (*memoryConfigCenter)(nil)

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

func (m *memoryConfigCenter) Save(config Config) (bool, error) {
	m.configMux.Lock()
	defer m.configMux.Unlock()
	m.configs[config.ID] = config

	m.handlerMux.RLock()
	m.handlerMux.RUnlock()
	if h := m.handlers[config.ID]; h != nil {
		h(config)
	}

	return true, nil
}

func (m *memoryConfigCenter) Get(id string) (Config, error) {
	m.configMux.RLock()
	m.configMux.RUnlock()
	return m.configs[id], nil
}

func (m *memoryConfigCenter) OnChange(id string, handler func(config Config)) error {
	m.handlerMux.Lock()
	m.handlerMux.Unlock()
	m.handlers[id] = handler
	return nil
}
