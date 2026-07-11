package server

import (
	"errors"
	"sync"
)

var (
	ErrServerAlreadyStarted = errors.New("server has already been started")
	ErrServerNotStarted     = errors.New("server has not been started")
	ErrServerStarting       = errors.New("server is still starting")
	ErrServerWaitConsumed   = errors.New("server wait result has already been consumed")
)

type componentState uint8

const (
	componentNew componentState = iota
	componentStarting
	componentRunning
	componentStopping
	componentStopped
)

// componentLifecycle protects the one-shot contract shared by concrete
// servers. A failed start is terminal because grpc.Server cannot be restarted.
type componentLifecycle struct {
	mu          sync.Mutex
	state       componentState
	started     bool
	waitClaimed bool
}

func (l *componentLifecycle) beginStart() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.state != componentNew {
		return ErrServerAlreadyStarted
	}
	l.state = componentStarting
	return nil
}

func (l *componentLifecycle) markStarted() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.started = true
	l.state = componentRunning
}

func (l *componentLifecycle) markStartFailed() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.state = componentStopped
}

func (l *componentLifecycle) claimWait() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.started {
		return ErrServerNotStarted
	}
	if l.waitClaimed {
		return ErrServerWaitConsumed
	}
	l.waitClaimed = true
	return nil
}

func (l *componentLifecycle) beginShutdown() (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	switch l.state {
	case componentNew, componentStopped, componentStopping:
		return false, nil
	case componentStarting:
		return false, ErrServerStarting
	case componentRunning:
		l.state = componentStopping
		return true, nil
	default:
		return false, nil
	}
}

func (l *componentLifecycle) markStopped() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.state = componentStopped
}
