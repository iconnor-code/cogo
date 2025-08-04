package core

import "google.golang.org/grpc"

type ServerOption func(IServer) error

type GrpcRegistar func(*grpc.Server, any)

type IServer interface {
	Start() error
	Stop() error
}
