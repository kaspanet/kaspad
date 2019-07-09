package locks

import (
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestPriorityMutex(t *testing.T) {
	mtx := NewPriorityMutex()
	sharedSlice := []int{}
	lowPriorityLockAcquired := false
	wg := sync.WaitGroup{}
	wg.Add(3)

	mtx.HighPriorityLock()
	go func() {
		mtx.LowPriorityLock()
		defer mtx.LowPriorityUnlock()
		sharedSlice = append(sharedSlice, 2)
		lowPriorityLockAcquired = true
		wg.Done()
	}()
	go func() {
		mtx.HighPriorityReadLock()
		defer mtx.HighPriorityReadUnlock()
		if lowPriorityLockAcquired {
			t.Errorf("LowPriorityLock unexpectedly released")
		}
		wg.Done()
	}()
	go func() {
		mtx.HighPriorityLock()
		defer mtx.HighPriorityUnlock()
		sharedSlice = append(sharedSlice, 1)
		if lowPriorityLockAcquired {
			t.Errorf("LowPriorityLock unexpectedly released")
		}
		wg.Done()
	}()
	time.Sleep(time.Second)
	mtx.HighPriorityUnlock()
	waitForWaitGroup(t, &wg, 2*time.Second)
	expectedSlice := []int{1, 2}
	if !reflect.DeepEqual(sharedSlice, expectedSlice) {
		t.Errorf("Expected the shared slice to be %d but got %d", expectedSlice, sharedSlice)
	}
}

func TestHighPriorityReadLock(t *testing.T) {
	mtx := NewPriorityMutex()
	wg := sync.WaitGroup{}
	wg.Add(2)
	mtx.LowPriorityLock()
	isReadLockHeld := false
	go func() {
		mtx.HighPriorityReadLock()
		defer mtx.HighPriorityReadUnlock()
		isReadLockHeld = true
		time.Sleep(1000 * time.Millisecond)
		isReadLockHeld = false
		wg.Done()
	}()
	go func() {
		time.Sleep(500 * time.Millisecond)
		mtx.HighPriorityReadLock()
		defer mtx.HighPriorityReadUnlock()
		if !isReadLockHeld {
			t.Errorf("expected another read lock to be held concurrently")
		}
		wg.Done()
	}()
	time.Sleep(time.Second)
	mtx.LowPriorityUnlock()
	waitForWaitGroup(t, &wg, 2*time.Second)
}

func waitForWaitGroup(t *testing.T, wg *sync.WaitGroup, timeout time.Duration) {
	doneWaiting := make(chan struct{})
	go func() {
		wg.Wait()
		doneWaiting <- struct{}{}
	}()
	select {
	case <-time.Tick(timeout):
		t.Fatalf("Unexpected timeout")
	case <-doneWaiting:
	}
}
