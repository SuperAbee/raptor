package servicecenter

var _ ServiceCenter = (*memoryServiceCenter)(nil)

// newMemoryServiceCenter for test
func newMemoryServiceCenter() ServiceCenter {
	return &memoryServiceCenter{}
}

type memoryServiceCenter struct {
	service Service
}

func (m *memoryServiceCenter) Register(param RegisterParam) (bool, error) {
	m.service = Service{
		Name:  param.ServiceName,
		Hosts: append(m.service.Hosts, Instance{
			Ip:      param.Ip,
			Port:    param.Port,
			Healthy: true,
		}),
	}
	return true, nil
}

func (m *memoryServiceCenter) GetService(name string) (Service, error) {
	if m.service.Name == name {
		return m.service, nil
	}
	return Service{
		Name: name,
		Hosts: []Instance{{Ip: "localhost", Port: 8878, Healthy: true}},
	}, nil
}
