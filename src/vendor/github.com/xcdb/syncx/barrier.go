package syncx

import "sync"

//Barrier enables multiple tasks to cooperatively work on an algorithm in parallel through multiple phases.
type Barrier struct {
	phase                 int
	participants, signals int
	cond                  *sync.Cond
	action                func()
}

//NewBarrier returns a new Barrier with participant count p and post-phase action a
//
//It panics if p is less than 0.
func NewBarrier(p int, a func()) *Barrier {
	if p < 0 {
		panic("syncx: NewBarrier p is less than 0")
	}
	cond := sync.NewCond(&sync.Mutex{})
	return &Barrier{
		phase:        1,
		participants: p,
		cond:         cond,
		action:       a,
	}
}

//Phase returns the number of the barrier's current phase.
func (b *Barrier) Phase() int {
	return b.phase
}

//Add adds delta, which may be negative, to the participants counter.
//If applying delta would cause the counter to go negative, Add panics.
func (b *Barrier) Add(delta int) {
	b.cond.L.Lock()
	defer b.cond.L.Unlock()
	if b.participants+delta < 0 {
		panic("syncx: negative Barrier participants counter")
	}
	b.participants += delta
	b.cond.Signal()
}

//SignalAndWait signals that a participant has reached the barrier and waits for all other participants to reach the barrier as well.
func (b *Barrier) SignalAndWait() {
	b.cond.L.Lock()
	defer b.cond.L.Unlock()
	b.signals++
	ph := b.phase
	for b.signals < b.participants && ph == b.phase {
		b.cond.Wait()
	}
	if ph == b.phase {
		if b.action != nil {
			b.action()
		}
		b.phase++
		b.signals = 0
	}
	b.cond.Broadcast()
}
