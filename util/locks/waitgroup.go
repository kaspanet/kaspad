package locks

import (
	"sync"
	"sync/atomic"
)

type waitGroup struct {
	addDoneCounter, waitingCounter int64
	waitLock, releaseWaitLock      sync.Mutex
	releaseWait, releaseDoneSpawn  chan struct{}
}

func newWaitGroup() *waitGroup {
	return &waitGroup{
		releaseWait:      make(chan struct{}),
		releaseDoneSpawn: make(chan struct{}),
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
			wg.releaseWaitLock.Lock()
			if atomic.LoadInt64(&wg.waitingCounter) > 0 {
				wg.releaseWait <- struct{}{}
				<-wg.releaseDoneSpawn
			}
			wg.releaseWaitLock.Unlock()
		})
	}
}

func (wg *waitGroup) wait() {
	wg.waitLock.Lock()
	defer wg.waitLock.Unlock()
	for atomic.LoadInt64(&wg.addDoneCounter) != 0 {
		atomic.AddInt64(&wg.waitingCounter, 1)
		<-wg.releaseWait
		atomic.AddInt64(&wg.waitingCounter, -1)
		wg.releaseDoneSpawn <- struct{}{}
	}
}
