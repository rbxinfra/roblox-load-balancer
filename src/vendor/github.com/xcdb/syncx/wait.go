package syncx

import (
	"context"
)

//WaitHandle represents any of AutoResetEvent, ManualResetEvent, Semaphore
type WaitHandle interface {
	ch() chan struct{}
}

var _Ø = make(chan struct{}, 1)

//WaitAny suspends execution of the calling goroutine until any handle receives a signal.
//
//Returns the array index of the handle that satisified the wait.
//If no handles are provided, returns -1.
func WaitAny(whs ...WaitHandle) int {
	if len(whs) > 8 {
		panic("Too many waithandles")
	}

	if len(whs) == 0 {
		return -1
	}

	cs := [8]chan struct{}{_Ø, _Ø, _Ø, _Ø, _Ø, _Ø, _Ø, _Ø}
	for i, wh := range whs {
		cs[i] = wh.ch()
	}

	select {
	case <-cs[0]:
		return 0
	case <-cs[1]:
		return 1
	case <-cs[2]:
		return 2
	case <-cs[3]:
		return 3
	case <-cs[4]:
		return 4
	case <-cs[5]:
		return 5
	case <-cs[6]:
		return 6
	case <-cs[7]:
		return 7
	}
}

//WaitAnyContext suspends execution of the calling goroutine until any handle receives a signal, or until the context is cancelled.
//
//Returns the array index of the handle that satisified the wait, or -1 and ctx.Err() if the context was cancelled.
//If no handles are provided, returns -1 with nil error.
func WaitAnyContext(ctx context.Context, whs ...WaitHandle) (int, error) {
	if len(whs) > 8 {
		panic("Too many waithandles")
	}

	if len(whs) == 0 {
		return -1, nil
	}

	cs := [8]chan struct{}{_Ø, _Ø, _Ø, _Ø, _Ø, _Ø, _Ø, _Ø}
	for i, wh := range whs {
		cs[i] = wh.ch()
	}

	select {
	case <-ctx.Done():
		return -1, ctx.Err()
	case <-cs[0]:
		return 0, nil
	case <-cs[1]:
		return 1, nil
	case <-cs[2]:
		return 2, nil
	case <-cs[3]:
		return 3, nil
	case <-cs[4]:
		return 4, nil
	case <-cs[5]:
		return 5, nil
	case <-cs[6]:
		return 6, nil
	case <-cs[7]:
		return 7, nil
	}
}

//WaitAll suspends execution of the calling goroutine until all handles have received a signal.
//
//Note that handles are not necessarily all in a signalled state at the same time...
//
//Returns true when all handles have satisified the wait.
func WaitAll(whs ...WaitHandle) bool {
	if len(whs) > 8 {
		panic("Too many waithandles")
	}

	if len(whs) == 0 {
		return true
	}

	cs := [8]chan struct{}{_Ø, _Ø, _Ø, _Ø, _Ø, _Ø, _Ø, _Ø}
	m := byte(0)
	for i, wh := range whs {
		cs[i] = wh.ch()
		m = m | (1 << uint(i))
	}

	var i uint
	for {
		select {
		case <-cs[0]:
			i = 0
		case <-cs[1]:
			i = 1
		case <-cs[2]:
			i = 2
		case <-cs[3]:
			i = 3
		case <-cs[4]:
			i = 4
		case <-cs[5]:
			i = 5
		case <-cs[6]:
			i = 6
		case <-cs[7]:
			i = 7
		}

		m = m &^ (1 << i)
		cs[i] = _Ø

		if m == 0 {
			return true
		}
	}
}

//WaitAllContext suspends execution of the calling goroutine until all handles have received a signal, or until the context is cancelled.
//
//Note that handles are not necessarily all in a signalled state at the same time...
//
//Returns true when all handles have satisified the wait, or false and ctx.Err() if the context was cancelled.
func WaitAllContext(ctx context.Context, whs ...WaitHandle) (bool, error) {
	if len(whs) > 8 {
		panic("Too many waithandles")
	}

	if len(whs) == 0 {
		return true, nil
	}

	cs := [8]chan struct{}{_Ø, _Ø, _Ø, _Ø, _Ø, _Ø, _Ø, _Ø}
	m := byte(0)
	for i, wh := range whs {
		cs[i] = wh.ch()
		m = m | (1 << uint(i))
	}

	var i uint
	for {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case <-cs[0]:
			i = 0
		case <-cs[1]:
			i = 1
		case <-cs[2]:
			i = 2
		case <-cs[3]:
			i = 3
		case <-cs[4]:
			i = 4
		case <-cs[5]:
			i = 5
		case <-cs[6]:
			i = 6
		case <-cs[7]:
			i = 7
		}

		m = m &^ (1 << i)
		cs[i] = _Ø

		if m == 0 {
			return true, nil
		}
	}
}
