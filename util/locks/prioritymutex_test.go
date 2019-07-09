package locks

import (
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestMutex(t *testing.T) {
	mtx := New()
	sharedSlice := []int{}
	lowPriorityLockReleased := false
	isReadLockHeld := false
	wg := sync.WaitGroup{}
	wg.Add(4)

	mtx.HighPriorityLock()
	go func() {
		mtx.LowPriorityLock()
		defer mtx.LowPriorityUnlock()
		sharedSlice = append(sharedSlice, 2)
		lowPriorityLockReleased = true
		wg.Done()
	}()
	go func() {
		mtx.HighPriorityReadLock()
		defer mtx.HighPriorityReadUnlock()
		if lowPriorityLockReleased {
			t.Errorf("LowPriorityLock unexpectedly released")
		}
		isReadLockHeld = true
		time.Sleep(time.Millisecond * 1000)
		isReadLockHeld = false
		wg.Done()
	}()
	go func() {
		time.Sleep(time.Millisecond * 500)
		mtx.HighPriorityReadLock()
		defer mtx.HighPriorityReadUnlock()
		if lowPriorityLockReleased {
			t.Errorf("LowPriorityLock unexpectedly released")
		}
		if !isReadLockHeld {
			t.Errorf("expected another read lock to be held concurrently")
		}
		wg.Done()
	}()
	go func() {
		mtx.HighPriorityLock()
		defer mtx.HighPriorityUnlock()
		sharedSlice = append(sharedSlice, 1)
		if lowPriorityLockReleased {
			t.Errorf("LowPriorityLock unexpectedly released")
		}
		wg.Done()
	}()
	time.Sleep(time.Second)
	mtx.HighPriorityUnlock()
	doneWaiting := make(chan struct{})
	go func() {
		wg.Wait()
		doneWaiting <- struct{}{}
	}()
	select {
	case <-time.Tick(2 * time.Second):
		t.Fatalf("Unexpected timeout")
	case <-doneWaiting:
	}
	expectedSlice := []int{1, 2}
	if !reflect.DeepEqual(sharedSlice, expectedSlice) {
		t.Errorf("Expected the shared slice to be %d but got %d", expectedSlice, sharedSlice)
	}
}
