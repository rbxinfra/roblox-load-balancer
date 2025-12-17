package syncx

import (
	"context"
	"sync"
)

//AutoResetEvent notifies a waiting goroutine that an event has occurred.
//
//Once it has been signaled, AutoResetEvent remains signaled until a single waiting goroutine is awoken, and then automatically returns to the non-signaled state.
//
//There is no guarantee that every call to Signal will wake a waiting goroutine.
//If two calls are too close together, so that the second call occurs before a goroutine has awoken, it is as if the second call did not happen.
//Also, if Set is called when there are no waiting goroutines, and e is already signaled, the call has no effect.
type AutoResetEvent struct {
	l sync.Mutex
	c chan struct{}
}

//NewAutoResetEvent returns a new AutoResetEvent with initial state s
func NewAutoResetEvent(s bool) *AutoResetEvent {
	e := AutoResetEvent{
		c: make(chan struct{}, 1),
	}
	if s {
		e.c <- struct{}{}
	}
	return &e
}

//Signal sets the state of e to signaled, waking a waiting goroutine.
func (e *AutoResetEvent) Signal() {
	e.l.Lock()
	if len(e.c) == 0 {
		e.c <- struct{}{}
	}
	e.l.Unlock()
}

//Reset sets the state of e to nonsignaled.
func (e *AutoResetEvent) Reset() {
	e.l.Lock()
	select {
	case <-e.c:
	default:
	}
	e.l.Unlock()
}

//Wait suspends execution of the calling goroutine until e receives a signal.
func (e *AutoResetEvent) Wait() {
	<-e.c
}

//WaitContext suspends execution of the calling goroutine until e receives a signal, or until the context is cancelled.
//The returned error is nil if e received a signal, or ctx.Err()
func (e *AutoResetEvent) WaitContext(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-e.c:
		return nil
	}
}

func (e *AutoResetEvent) ch() chan struct{} {
	return e.c
}
