package core

import "google.golang.org/grpc"

// IRPCClient owns shared gRPC connections to downstream services.
// Implementations must reuse connections when possible and release all owned
// connections when Close is called.
type IRPCClient interface {
	Conn(service string) (*grpc.ClientConn, error)
	Close() error
}
