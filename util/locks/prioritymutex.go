package locks

import (
	"sync"
)

// PriorityMutex is a read-write mutex with an additional low
// priority lock. It's implemented with the following
// components:
//  * Data mutex: The actual lock on the data structure. Its
//    type is sync.RWMutex for its high priority read lock.
//  * High priority waiting group: A waiting group that is being
//    increased every time a high priority lock (read or write)
//    is acquired, and decreased every time a high priority lock is
//    unlocked. Low priority locks can start being held only
//    when the waiting group is empty.
//	* Low priority mutex: This mutex ensures that when the
//    waiting group is empty, only one low priority lock
//    will be able to lock the data mutex.

// PriorityMutex implements a lock with three priorities:
//	* High priority write lock - locks the mutex with the highest priority.
//  * High priority read lock - locks the mutex with lower priority than
//    the high priority write lock. Can be held concurrently with other
//    with other read locks.
//  * Low priority write lock - locks the mutex with lower priority then
//	  the read lock.
type PriorityMutex struct {
	dataMutex           sync.RWMutex
	highPriorityWaiting sync.WaitGroup
	lowPriorityMutex    sync.Mutex
}

func NewPriorityMutex() *PriorityMutex {
	lock := PriorityMutex{
		highPriorityWaiting: sync.WaitGroup{},
	}
	return &lock
}

// LowPriorityLock will acquire a low-priority lock.
func (mtx *PriorityMutex) LowPriorityLock() {
	mtx.lowPriorityMutex.Lock()
	mtx.highPriorityWaiting.Wait()
	mtx.dataMutex.Lock()
}

// LowPriorityUnlock will unlock the low-priority lock
func (mtx *PriorityMutex) LowPriorityUnlock() {
	mtx.dataMutex.Unlock()
	mtx.lowPriorityMutex.Unlock()
}

// HighPriorityLock will acquire a high-priority lock.
func (mtx *PriorityMutex) HighPriorityLock() {
	mtx.highPriorityWaiting.Add(1)
	mtx.dataMutex.Lock()
}

// HighPriorityUnlock will unlock the high-priority lock
func (mtx *PriorityMutex) HighPriorityUnlock() {
	mtx.dataMutex.Unlock()
	mtx.highPriorityWaiting.Done()
}

// HighPriorityReadLock will acquire a high-priority read
// lock.
func (mtx *PriorityMutex) HighPriorityReadLock() {
	mtx.highPriorityWaiting.Add(1)
	mtx.dataMutex.RLock()
}

// HighPriorityUnlock will unlock the high-priority read
// lock
func (mtx *PriorityMutex) HighPriorityReadUnlock() {
	mtx.highPriorityWaiting.Done()
	mtx.dataMutex.RUnlock()
}
