package locks

import (
	"sync"
	"sync/atomic"
)

type waitGroup struct {
	counter     int64
	waitLock    sync.Mutex
	syncChannel chan struct{}
}

func newWaitGroup() *waitGroup {
	return &waitGroup{
		waitLock:    sync.Mutex{},
		syncChannel: make(chan struct{}),
	}
}

func (wg *waitGroup) add() {
	atomic.AddInt64(&wg.counter, 1)
}

func (wg *waitGroup) done() {
	counter := atomic.AddInt64(&wg.counter, -1)
	if counter < 0 {
		panic("negative values for wg.counter are not allowed. This was likely caused by calling done() before add()")
	}
	if atomic.LoadInt64(&wg.counter) == 0 {
		spawn(func() {
			wg.syncChannel <- struct{}{}
		})
	}
}

func (wg *waitGroup) wait() {
	wg.waitLock.Lock()
	defer wg.waitLock.Unlock()
	for atomic.LoadInt64(&wg.counter) != 0 {
		<-wg.syncChannel
	}
}
