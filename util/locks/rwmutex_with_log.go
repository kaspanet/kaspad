package locks

import (
	"github.com/kaspanet/kaspad/logger"
	"sync"
)

const rwMutexWithLogFileName = "rwmutex_with_log.go"

// RWMutexWithLog is a wrapper for sync.RWMutex that logs
// any lock and unlock.
type RWMutexWithLog struct {
	sync.RWMutex
}

// Lock locks RWMutexWithLog underlying sync.RWMutex
func (rwm *RWMutexWithLog) Lock() {
	log.Debugf("RWMutexWithLog.Lock():%s", logger.NewLogClosure(func() string {
		return goroutineIDAndCallerToMutex(rwMutexWithLogFileName)
	}))
	rwm.RWMutex.Lock()
}

// Unlock unlocks RWMutexWithLog underlying sync.RWMutex
func (rwm *RWMutexWithLog) Unlock() {
	log.Debugf("RWMutexWithLog.Unlock():%s", logger.NewLogClosure(func() string {
		return goroutineIDAndCallerToMutex(rwMutexWithLogFileName)
	}))
	rwm.RWMutex.Unlock()
}

// RLock read-locks RWMutexWithLog underlying sync.RWMutex
func (rwm *RWMutexWithLog) RLock() {
	log.Debugf("RWMutexWithLog.RLock():%s", logger.NewLogClosure(func() string {
		return goroutineIDAndCallerToMutex(rwMutexWithLogFileName)
	}))
	rwm.RWMutex.RLock()
}

// RUnlock read-unlocks RWMutexWithLog underlying sync.RWMutex
func (rwm *RWMutexWithLog) RUnlock() {
	log.Debugf("RWMutexWithLog.RUnlock():%s", logger.NewLogClosure(func() string {
		return goroutineIDAndCallerToMutex(rwMutexWithLogFileName)
	}))
	rwm.RWMutex.RUnlock()
}
