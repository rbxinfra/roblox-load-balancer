package syncx

import (
	"context"
	"sync"
)

//ManualResetEvent notifies one or more waiting goroutines that an event has occurred.
//
//Once it has been signaled, ManualResetEvent remains signaled until it is manually reset.
//When signaled, all waiting goroutines are released, and all calls to Wait return immediately.
type ManualResetEvent struct {
	l sync.Mutex
	c chan struct{}
}

//NewManualResetEvent returns a new ManualResetEvent with initial state s
func NewManualResetEvent(s bool) *ManualResetEvent {
	e := ManualResetEvent{
		c: make(chan struct{}, 1),
	}
	if s {
		close(e.c)
	}
	return &e
}

//Signal sets the state of e to signaled, waking one or more waiting goroutines.
func (e *ManualResetEvent) Signal() {
	e.l.Lock()
	select {
	case <-e.c: //ch is closed
	default:
		close(e.c)
	}
	e.l.Unlock()
}

//Reset sets the state of e to nonsignaled.
func (e *ManualResetEvent) Reset() {
	e.l.Lock()
	select {
	case <-e.c: //ch is closed
		e.c = make(chan struct{}, 1)
	default:
	}
	e.l.Unlock()
}

//Wait suspends execution of the calling goroutine until e receives a signal.
func (e *ManualResetEvent) Wait() {
	<-e.c
}

//WaitContext suspends execution of the calling goroutine until e receives a signal, or until the context is cancelled.
//The returned error is nil if e received a signal, or ctx.Err()
func (e *ManualResetEvent) WaitContext(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-e.c:
		return nil
	}
}

func (e *ManualResetEvent) ch() chan struct{} {
	return e.c
}
