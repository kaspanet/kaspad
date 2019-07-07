package prioritylock

import (
	"sync"
	"testing"
	"time"
)

func TestMutex(t *testing.T) {
	mtx := New()
	mtx.HighPriorityLock()
	lowPriorityLockReleased := false
	isReadLockHeld := false
	wg := sync.WaitGroup{}
	wg.Add(4)
	go func() {
		mtx.LowPriorityLock()
		defer mtx.LowPriorityUnlock()
		lowPriorityLockReleased = true
		wg.Done()
	}()
	go func() {
		mtx.HighPriorityReadLock()
		defer mtx.HighPriorityReadUnlock()
		isReadLockHeld = true
		time.Sleep(time.Millisecond * 1000)
		isReadLockHeld = false
		wg.Done()
	}()
	go func() {
		time.Sleep(time.Millisecond * 500)
		mtx.HighPriorityReadLock()
		defer mtx.HighPriorityReadUnlock()
		if !isReadLockHeld {
			t.Errorf("expected another read lock to be held concurrently")
		}
		wg.Done()
	}()
	go func() {
		mtx.HighPriorityLock()
		defer mtx.HighPriorityUnlock()
		if lowPriorityLockReleased {
			t.Errorf("LowPriorityLock unexpectedly released")
		}
		wg.Done()
	}()
	time.Sleep(time.Second)
	mtx.HighPriorityUnlock()
	wg.Wait()
}
