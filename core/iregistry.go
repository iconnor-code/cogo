package core

import "context"

type RegistryOption func(IRegistry) error

type IRegistry interface {
	Register(ctx context.Context, serverName string, serverAddr string) error
	Deregister(ctx context.Context, serverName string, serverAddr string) error
}
