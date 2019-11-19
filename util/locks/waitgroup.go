package locks

import (
	"sync"
	"sync/atomic"
)

type waitGroup struct {
	addDoneCounter, waitingCounter int64
	waitLock                       sync.Mutex
	syncChannel                    chan struct{}
}

func newWaitGroup() *waitGroup {
	return &waitGroup{
		waitLock:    sync.Mutex{},
		syncChannel: make(chan struct{}),
	}
}

func (wg *waitGroup) add() {
	atomic.AddInt64(&wg.addDoneCounter, 1)
}

func (wg *waitGroup) done() {
	counter := atomic.AddInt64(&wg.addDoneCounter, -1)
	if counter < 0 {
		panic("negative values for wg.addDoneCounter are not allowed. This was likely caused by calling done() before add()")
	}
	if atomic.LoadInt64(&wg.addDoneCounter) == 0 && atomic.LoadInt64(&wg.waitingCounter) > 0 {
		spawn(func() {
			if atomic.LoadInt64(&wg.waitingCounter) > 0 {
				wg.syncChannel <- struct{}{}
			}
		})
	}
}

func (wg *waitGroup) wait() {
	atomic.AddInt64(&wg.waitingCounter, 1)
	defer atomic.AddInt64(&wg.waitingCounter, -1)
	wg.waitLock.Lock()
	defer wg.waitLock.Unlock()
	for atomic.LoadInt64(&wg.addDoneCounter) != 0 {
		<-wg.syncChannel
	}
}
