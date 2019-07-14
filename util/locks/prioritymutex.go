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
//    incremented every time a high priority lock (read or write)
//    is acquired, and decremented every time a high priority lock is
//    unlocked. Low priority locks can start being held only
//    when the waiting group is empty.
//  * Low priority mutex: This mutex ensures that when the
//    waiting group is empty, only one low priority lock
//    will be able to lock the data mutex.

// PriorityMutex implements a lock with three priorities:
//  * High priority write lock - locks the mutex with the highest priority.
//  * High priority read lock - locks the mutex with lower priority than
//    the high priority write lock. Can be held concurrently with other
//    with other read locks.
//  * Low priority write lock - locks the mutex with lower priority then
//	  the read lock.
type PriorityMutex struct {
	dataMutex           sync.RWMutex
	highPriorityWaiting *waitGroup
}

func NewPriorityMutex() *PriorityMutex {
	lock := PriorityMutex{
		highPriorityWaiting: newWaitGroup(),
	}
	return &lock
}

// LowPriorityWriteLock acquires a low-priority write lock.
func (mtx *PriorityMutex) LowPriorityWriteLock() {
	mtx.highPriorityWaiting.wait()
	mtx.dataMutex.Lock()
}

// LowPriorityWriteUnlock unlocks the low-priority write lock
func (mtx *PriorityMutex) LowPriorityWriteUnlock() {
	mtx.dataMutex.Unlock()
}

// HighPriorityWriteLock acquires a high-priority write lock.
func (mtx *PriorityMutex) HighPriorityWriteLock() {
	mtx.highPriorityWaiting.add()
	mtx.dataMutex.Lock()
}

// HighPriorityWriteUnlock unlocks the high-priority write lock
func (mtx *PriorityMutex) HighPriorityWriteUnlock() {
	mtx.dataMutex.Unlock()
	mtx.highPriorityWaiting.done()
}

// HighPriorityReadLock acquires a high-priority read
// lock.
func (mtx *PriorityMutex) HighPriorityReadLock() {
	mtx.highPriorityWaiting.add()
	mtx.dataMutex.RLock()
}

// HighPriorityWriteUnlock unlocks the high-priority read
// lock
func (mtx *PriorityMutex) HighPriorityReadUnlock() {
	mtx.highPriorityWaiting.done()
	mtx.dataMutex.RUnlock()
}
