package core

import "context"

// Server is a long-running process component whose startup, runtime, and
// shutdown failures are observable by its owner.
type Server interface {
	Start(context.Context) error
	Wait() error
	Shutdown(context.Context) error
}
