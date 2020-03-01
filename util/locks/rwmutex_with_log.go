package locks

import (
	"sync"
)

// RWMutexWithLog is a wrapper for sync.RWMutex that logs
// any lock and unlock.
type RWMutexWithLog struct {
	sync.RWMutex
}

// Lock locks RWMutexWithLog underlying sync.RWMutex
func (rwm *RWMutexWithLog) Lock() {
	log.Debugf("RWMutexWithLog.Lock():%s", goroutineIDAndCallerToMutex())
	rwm.RWMutex.Lock()
}

// Unlock unlocks RWMutexWithLog underlying sync.RWMutex
func (rwm *RWMutexWithLog) Unlock() {
	log.Debugf("RWMutexWithLog.Unlock():%s", goroutineIDAndCallerToMutex())
	rwm.RWMutex.Unlock()
}

// RLock read-locks RWMutexWithLog underlying sync.RWMutex
func (rwm *RWMutexWithLog) RLock() {
	log.Debugf("RWMutexWithLog.RLock():%s", goroutineIDAndCallerToMutex())
	rwm.RWMutex.RLock()
}

// RUnlock read-unlocks RWMutexWithLog underlying sync.RWMutex
func (rwm *RWMutexWithLog) RUnlock() {
	log.Debugf("RWMutexWithLog.RUnlock():%s", goroutineIDAndCallerToMutex())
	rwm.RWMutex.RUnlock()
}
