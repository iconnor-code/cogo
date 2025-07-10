package core

type ServerOption func(IServer) error

type IServer interface {
	Start() error
	Stop() error
}
