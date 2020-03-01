package locks

import (
	"sync"
)

// RWMutexWithLog is a wrapper for sync.RWMutex that logs
// any lock and unlock.
type RWMutexWithLog struct {
	sync.RWMutex
}

func (rwm *RWMutexWithLog) Lock() {
	log.Debugf("RWMutexWithLog.Lock():%s", goroutineIDAndCallerToMutex())
	rwm.RWMutex.Lock()
}

func (rwm *RWMutexWithLog) Unlock() {
	log.Debugf("RWMutexWithLog.Unlock():%s", goroutineIDAndCallerToMutex())
	rwm.RWMutex.Unlock()
}

func (rwm *RWMutexWithLog) RLock() {
	log.Debugf("RWMutexWithLog.RLock():%s", goroutineIDAndCallerToMutex())
	rwm.RWMutex.RLock()
}

func (rwm *RWMutexWithLog) RUnlock() {
	log.Debugf("RWMutexWithLog.RUnlock():%s", goroutineIDAndCallerToMutex())
	rwm.RWMutex.RUnlock()
}
