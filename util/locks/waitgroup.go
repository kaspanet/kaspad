package locks

import (
	"sync"
	"sync/atomic"
)

// waitGroup is a type that implements the same API
// as sync.WaitGroup but allows concurrent calls to
// add() and wait().
type waitGroup struct {
	counter, isReleaseWaitWaiting          int64
	mainWaitLock, isReleaseWaitWaitingLock sync.Mutex
	releaseWait, releaseDoneSpawn          chan struct{}
	id                                     uint64
}

func newWaitGroup() *waitGroup {
	return &waitGroup{
		releaseWait:      make(chan struct{}),
		releaseDoneSpawn: make(chan struct{}),
	}
}

func (wg *waitGroup) add(delta int64) {
	atomic.AddInt64(&wg.counter, delta)
}

func (wg *waitGroup) done() {
	counter := atomic.AddInt64(&wg.counter, -1)
	if counter < 0 {
		panic("negative values for wg.counter are not allowed. This was likely caused by calling done() before add()")
	}

	// To avoid a situation where a struct is
	// being sent to wg.releaseWait while there
	// are no listeners to the channel (which will
	// cause the goroutine to hang for eternity),
	// we check wg.isReleaseWaitWaiting to see
	// if there is a listener to wg.releaseWait.
	if atomic.LoadInt64(&wg.counter) == 0 && atomic.LoadInt64(&wg.isReleaseWaitWaiting) == 1 {
		spawn(func() {
			wg.isReleaseWaitWaitingLock.Lock()
			if atomic.LoadInt64(&wg.isReleaseWaitWaiting) == 1 {
				wg.releaseWait <- struct{}{}
				<-wg.releaseDoneSpawn
			}
			wg.isReleaseWaitWaitingLock.Unlock()
		})
	} else {
	}
}

func (wg *waitGroup) wait() {
	wg.mainWaitLock.Lock()
	defer wg.mainWaitLock.Unlock()
	wg.isReleaseWaitWaitingLock.Lock()
	for atomic.LoadInt64(&wg.counter) != 0 {
		atomic.StoreInt64(&wg.isReleaseWaitWaiting, 1)
		wg.isReleaseWaitWaitingLock.Unlock()
		<-wg.releaseWait
		atomic.StoreInt64(&wg.isReleaseWaitWaiting, 0)
		wg.releaseDoneSpawn <- struct{}{}
		wg.isReleaseWaitWaitingLock.Lock()
	}
	wg.isReleaseWaitWaitingLock.Unlock()
}
