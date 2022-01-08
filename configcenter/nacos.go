package configcenter

var _ ConfigCenter = (*nacosConfigCenter)(nil)

func newNacosConfigCenter() ConfigCenter {
	return &nacosConfigCenter{}
}

type nacosConfigCenter struct {
}

func (n *nacosConfigCenter) Get(name string) Config {
	return Config{}
}
