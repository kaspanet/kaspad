package locks

import (
	"sync"
)

// MutexWithLog is a wrapper for sync.Mutex that logs
// any lock and unlock.
type MutexWithLog struct {
	sync.Mutex
}

// Lock locks MutexWithLog underlying sync.Mutex
func (rwm *MutexWithLog) Lock() {
	log.Debugf("MutexWithLog.Lock():%s", goroutineIDAndCallerToMutex())
	rwm.Mutex.Lock()
}

// Unlock unlocks MutexWithLog underlying sync.Mutex
func (rwm *MutexWithLog) Unlock() {
	log.Debugf("MutexWithLog.Unlock():%s", goroutineIDAndCallerToMutex())
	rwm.Mutex.Unlock()
}
