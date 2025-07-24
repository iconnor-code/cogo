package discovery

// 服务实例
type ServiceInstance struct {
	ID   string
	Name string
	Addr string
	Port int
	Tags []string
	Meta map[string]string
}

func (si *ServiceInstance) GetName() string {
	return si.Name
}
