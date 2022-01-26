package proto

// ServiceCenter 张晨光要求姚新聪提供的服务中心接口
type ServiceCenter interface {
	// GetTargetServices 获取被调服务信息
	GetTargetServices(name string) []Service
	// Register 注册当前调度器
	Register(Service Service) error
	// GetServices 获取集群中所有调度器信息
	GetServices() []Service
}

type Service struct {
	Name  string     `json:"name"`
	Hosts []Instance `json:"hosts"`
}

type Instance struct {
	Ip      string `json:"ip"`
	Port    uint64 `json:"port"`
	Healthy bool   `json:"healthy"`
}
