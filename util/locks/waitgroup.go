package locks

import (
	"sync"
	"sync/atomic"
)

type waitGroup struct {
	counter  int64
	waitCond *sync.Cond
}

func newWaitGroup() *waitGroup {
	return &waitGroup{
		waitCond: sync.NewCond(&sync.Mutex{}),
	}
}

func (wg *waitGroup) add() {
	atomic.AddInt64(&wg.counter, 1)
}

func (wg *waitGroup) done() {
	counter := atomic.AddInt64(&wg.counter, -1)
	if counter == 0 {
		wg.waitCond.Signal()
	}
	if counter < 0 {
		panic("negative values for wg.counter are not allowed. It's probably because done() was called before add()")
	}
}

func (wg *waitGroup) wait() {
	wg.waitCond.L.Lock()
	defer wg.waitCond.L.Unlock()
	for wg.counter != 0 {
		wg.waitCond.Wait()
	}
}
