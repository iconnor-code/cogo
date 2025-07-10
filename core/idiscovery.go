package core

type IDiscovery interface {
	GetGrpcClientConn(serverName string) (any, error)
}

type DiscoveryOption func(d IDiscovery) error
