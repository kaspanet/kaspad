package prioritylock

import (
	"sync"
)

// Mutex implements a lock with three priorities:
// 	* High priority write lock - locks the mutex with the highest priority.
//  * High priority read lock - locks the mutex with lower priority than
//    the high priority write lock. Can be help concurrently with other
//    with other read locks.
//  * Low priority read lock - locks the mutex with lower priority then
//	  the read lock.
type Mutex struct {
	dataMutex           sync.RWMutex
	lowPriorityMutex    sync.Mutex
	highPriorityWaiting sync.WaitGroup
}

func New() *Mutex {
	lock := Mutex{
		highPriorityWaiting: sync.WaitGroup{},
	}
	return &lock
}

// LowPriorityLock will acquire a low-priority lock
// it must wait until both low priority and all high priority lock holders are released.
func (mtx *Mutex) LowPriorityLock() {
	mtx.lowPriorityMutex.Lock()
	mtx.highPriorityWaiting.Wait()
	mtx.dataMutex.Lock()
}

// LowPriorityUnlock will unlock the low-priority lock
func (mtx *Mutex) LowPriorityUnlock() {
	mtx.dataMutex.Unlock()
	mtx.lowPriorityMutex.Unlock()
}

// HighPriorityLock will acquire a high-priority lock
// it must still wait until a low-priority lock has been released.
func (mtx *Mutex) HighPriorityLock() {
	mtx.highPriorityWaiting.Add(1)
	mtx.dataMutex.Lock()
}

// HighPriorityUnlock will unlock the high-priority lock
func (mtx *Mutex) HighPriorityUnlock() {
	mtx.dataMutex.Unlock()
	mtx.highPriorityWaiting.Done()
}

func (mtx *Mutex) HighPriorityReadLock() {
	mtx.highPriorityWaiting.Add(1)
	mtx.dataMutex.RLock()
}

func (mtx *Mutex) HighPriorityReadUnlock() {
	mtx.highPriorityWaiting.Done()
	mtx.dataMutex.RUnlock()
}
