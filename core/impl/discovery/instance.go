package discovery

import "fmt"

type ServerInstance struct {
	ID      string
	Name    string
	Address string
	Port    int
	Tags    []string
	Meta    map[string]string
}

func (si *ServerInstance) GetName() string {
	return si.Name
}

func (si *ServerInstance) GetAddr() string {
	return fmt.Sprintf("%v:%v", si.Address, si.Port)
}
