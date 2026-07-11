package core

import "context"

// Server is a one-shot long-running component whose startup, runtime, and
// shutdown failures are observable by its owner. Start may be called once,
// Wait consumes one terminal result, and Shutdown must cause Wait to return.
type Server interface {
	Start(context.Context) error
	Wait() error
	Shutdown(context.Context) error
}
