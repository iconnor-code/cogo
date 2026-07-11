package core

import "context"

type IRegistry interface {
	Register(ctx context.Context) error
	DeRegister(ctx context.Context) error
}
