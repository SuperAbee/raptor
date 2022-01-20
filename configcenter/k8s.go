package configcenter

var _ ConfigCenter = (*k8sConfigCenter)(nil)

func newK8sConfigCenter() ConfigCenter {
	return &k8sConfigCenter{}
}

type k8sConfigCenter struct {
}

func (k *k8sConfigCenter) Save(config Config) (bool, error) {
	return true, nil
}

func (k *k8sConfigCenter) Get(id string) (Config, error) {
	return Config{}, nil
}

func (k *k8sConfigCenter) OnChange(id string, handler func(config Config)) error {
	return nil
}
