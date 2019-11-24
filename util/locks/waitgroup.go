package locks

import (
	"sync"
	"sync/atomic"
)

// waitGroup is a type that implements the same API
// as sync.WaitGroup but allows concurrent calls to
// add() and wait().
//
// The wait group maintains a counter that can be
// increased by delta by using the waitGroup.add
// method, and decreased by 1 by using the
// waitGroup.done method.
// The general idea of the waitGroup.wait method
// is to block the current goroutine until the
// counter is set to 0. This is how it's implemented:
//
// 1) The waitGroup.wait method is checking if waitGroup.counter
//   is 0. If it's the case the function returns. otherwise,
//   it sets the flag waitGroup.isReleaseWaitWaiting to 1 so
//   that there's a pending wait function, and waits for a signal
//   from the channel waitGroup.relaseWait (waitGroup.isReleaseWaitWaiting
//   is set to 1 wrapped with waitGroup.isReleaseWaitWaitingLock to
//   synchronize with the reader from waitGroup.done).
//
// 2) When waitGroup.done is called, it checks if waitGroup.counter
//    is 0.
//
// 3) If waitGroup.counter is 0, it also checks if there's any pending
//    wait function by checking if wg.isReleaseWaitWaiting is 1, and if
//    this is the case, it sends a signal to release the pending wait
//    function, and then waits for a signal from waitGroup.releaseDone,
//    and when the signal is received, the function returns.
//    This step is wrapped with isReleaseWaitWaitingLock for two reasons:
//    a) Prevent a situation where waitGroup.wait goroutine yields just
//       before it sets wg.isReleaseWaitWaiting to 1, and then
//       waitGroup.done will exit the function without sending the signal
//       to waitGroup.wait.
//    b) Prevent two waitGroup.done send concurrently a signal to the
//       channel wg.releaseWait and making one of them hang forever.
//
// 4) After the waitGroup.wait is released, it sets
//    waitGroup.isReleaseWaitWaiting to 0, and sends
//    a signal to wg.releaseDone and go back to step 1.
//
// The waitGroup.wait is wrapped with waitGroup.mainWaitLock. It
// is used to enable multiple waits pending for the counter to be
// set to zero. This will cause a situation when one wait function
// will return, the other waits that are pending to waitGroup.mainWaitLock
// will immediately return as well. Without that lock, any call
// to waitGroup.wait will wait to its own signal from waitGroup.releaseWait
// which means that for n waits to be unblocked, the counter has to be set
// to 0 n times.
type waitGroup struct {
	counter, isReleaseWaitWaiting          int64
	mainWaitLock, isReleaseWaitWaitingLock sync.Mutex
	releaseWait, releaseDone               chan struct{}
}

func newWaitGroup() *waitGroup {
	return &waitGroup{
		releaseWait: make(chan struct{}),
		releaseDone: make(chan struct{}),
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
	if atomic.LoadInt64(&wg.counter) == 0 {
		wg.isReleaseWaitWaitingLock.Lock()
		if atomic.LoadInt64(&wg.isReleaseWaitWaiting) == 1 {
			wg.releaseWait <- struct{}{}
			<-wg.releaseDone
		}
		wg.isReleaseWaitWaitingLock.Unlock()
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
		wg.releaseDone <- struct{}{}
		wg.isReleaseWaitWaitingLock.Lock()
	}
	wg.isReleaseWaitWaitingLock.Unlock()
}
