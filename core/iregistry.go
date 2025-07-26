package core

import "context"

type RegistryOption func(IRegistry) error

type IRegistry interface {
	Register(ctx context.Context) error
	DeRegister(ctx context.Context) error
}
