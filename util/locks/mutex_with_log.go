package locks

import (
	"github.com/kaspanet/kaspad/logger"
	"sync"
)

const mutexWithLogFileName = "mutex_with_log.go"

// MutexWithLog is a wrapper for sync.Mutex that logs
// any lock and unlock.
type MutexWithLog struct {
	sync.Mutex
}

// Lock locks MutexWithLog underlying sync.Mutex
func (m *MutexWithLog) Lock() {
	log.Debugf("MutexWithLog.Lock():%s", logger.NewLogClosure(func() string {
		return goroutineIDAndCallerToMutex(mutexWithLogFileName)
	}))
	m.Mutex.Lock()
}

// Unlock unlocks MutexWithLog underlying sync.Mutex
func (m *MutexWithLog) Unlock() {
	log.Debugf("MutexWithLog.Unlock():%s", logger.NewLogClosure(func() string {
		return goroutineIDAndCallerToMutex(mutexWithLogFileName)
	}))
	m.Mutex.Unlock()
}
