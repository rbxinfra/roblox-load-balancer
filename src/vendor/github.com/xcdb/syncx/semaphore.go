package syncx

import (
	"context"
)

//Semaphore limits the number of goroutines that can access a resource or pool of resources concurrently.
type Semaphore struct {
	c chan struct{}
}

//NewSemaphore returns a new Semaphore with count c
//
//It panics if c is less than 1.
func NewSemaphore(c int) *Semaphore {
	if c < 1 {
		panic("syncx: NewSemaphore c is less than 1")
	}
	s := Semaphore{
		c: make(chan struct{}, c),
	}
	for i := 0; i < c; i++ {
		s.c <- struct{}{}
	}
	return &s
}

//Release exits s, waking a waiting goroutine.
//
//If the Semaphore reaches maximum capacity, further calls to Release are ignored.
func (s *Semaphore) Release() {
	if len(s.c) < cap(s.c) {
		s.c <- struct{}{}
	}
}

//Wait suspends execution of the calling goroutine until it can enter s.
func (s *Semaphore) Wait() {
	<-s.c
}

//WaitContext suspends execution of the calling goroutine until e receives a signal, or until the context is cancelled.
//
//The returned error is nil if e received a signal, or ctx.Err()
func (s *Semaphore) WaitContext(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-s.c:
		return nil
	}
}

func (s *Semaphore) ch() chan struct{} {
	return s.c
}
