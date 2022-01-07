package servicecenter

var _ ServiceCenter = (*k8sServiceCenter)(nil)

func newK8sServiceCenter() ServiceCenter {
	return &k8sServiceCenter{}
}

type k8sServiceCenter struct {
}

func (n *k8sServiceCenter) GetService(name string) Service {
	return Service{}
}
