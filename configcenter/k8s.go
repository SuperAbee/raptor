package configcenter

var _ ConfigCenter = (*k8sConfigCenter)(nil)

func newK8sConfigCenter() ConfigCenter {
	return &k8sConfigCenter{}
}

type k8sConfigCenter struct {
}

func (n *k8sConfigCenter) Get(name string) Config {
	return Config{}
}
