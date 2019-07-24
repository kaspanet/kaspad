// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package locks

import (
	"sync/atomic"
	"testing"
)

type syncWgCompatible struct {
	*waitGroup
}

func (swg *syncWgCompatible) Add(delta int) {
	for i := 0; i < delta; i++ {
		swg.add()
	}
}

func (swg *syncWgCompatible) Done() {
	swg.done()
}

func (swg *syncWgCompatible) Wait() {
	swg.wait()
}

func newSyncWgCompatible() *syncWgCompatible {
	return &syncWgCompatible{
		waitGroup: newWaitGroup(),
	}
}

func testWaitGroup(t *testing.T, wg1 *syncWgCompatible, wg2 *syncWgCompatible) {
	n := 16
	wg1.Add(n)
	wg2.Add(n)
	exited := make(chan bool, n)
	for i := 0; i != n; i++ {
		go func(i int) {
			wg1.Done()
			wg2.Wait()
			exited <- true
		}(i)
	}
	wg1.Wait()
	for i := 0; i != n; i++ {
		select {
		case <-exited:
			t.Fatal("WaitGroup released group too soon")
		default:
		}
		wg2.Done()
	}
	for i := 0; i != n; i++ {
		<-exited // Will block if barrier fails to unlock someone.
	}
}

func TestWaitGroup(t *testing.T) {
	wg1 := newSyncWgCompatible()
	wg2 := newSyncWgCompatible()

	// Run the same test a few times to ensure barrier is in a proper state.
	for i := 0; i != 8; i++ {
		testWaitGroup(t, wg1, wg2)
	}
}

func TestWaitGroupMisuse(t *testing.T) {
	defer func() {
		err := recover()
		if err != "negative values for wg.counter are not allowed. This was likely caused by calling done() before add()" {
			t.Fatalf("Unexpected panic: %#v", err)
		}
	}()
	wg := newSyncWgCompatible()
	wg.Add(1)
	wg.Done()
	wg.Done()
	t.Fatal("Should panic")
}

func TestAddAfterWait(t *testing.T) {
	wg := newSyncWgCompatible()
	wg.add()
	syncChan := make(chan struct{})
	go func() {
		syncChan <- struct{}{}
		wg.wait()
		syncChan <- struct{}{}
	}()
	<-syncChan
	wg.add()
	wg.done()
	wg.done()
	<-syncChan
}

func TestWaitGroupRace(t *testing.T) {
	// Run this test for about 1ms.
	for i := 0; i < 1000; i++ {
		wg := newSyncWgCompatible()
		n := new(int32)
		// spawn goroutine 1
		wg.Add(1)
		go func() {
			atomic.AddInt32(n, 1)
			wg.Done()
		}()
		// spawn goroutine 2
		wg.Add(1)
		go func() {
			atomic.AddInt32(n, 1)
			wg.Done()
		}()
		// Wait for goroutine 1 and 2
		wg.Wait()
		if atomic.LoadInt32(n) != 2 {
			t.Fatal("Spurious wakeup from Wait")
		}
	}
}

func TestWaitGroupAlign(t *testing.T) {
	type X struct {
		x  byte
		wg *syncWgCompatible
	}
	x := X{wg: newSyncWgCompatible()}
	x.wg.Add(1)
	go func(x *X) {
		x.wg.Done()
	}(&x)
	x.wg.Wait()
}

func BenchmarkWaitGroupUncontended(b *testing.B) {
	type PaddedWaitGroup struct {
		*syncWgCompatible
		pad [128]uint8
	}
	b.RunParallel(func(pb *testing.PB) {
		wg := PaddedWaitGroup{
			syncWgCompatible: newSyncWgCompatible(),
		}
		for pb.Next() {
			wg.Add(1)
			wg.Done()
			wg.Wait()
		}
	})
}

func benchmarkWaitGroupAddDone(b *testing.B, localWork int) {
	wg := newSyncWgCompatible()
	b.RunParallel(func(pb *testing.PB) {
		foo := 0
		for pb.Next() {
			wg.Add(1)
			for i := 0; i < localWork; i++ {
				foo *= 2
				foo /= 2
			}
			wg.Done()
		}
		_ = foo
	})
}

func BenchmarkWaitGroupAddDone(b *testing.B) {
	benchmarkWaitGroupAddDone(b, 0)
}

func BenchmarkWaitGroupAddDoneWork(b *testing.B) {
	benchmarkWaitGroupAddDone(b, 100)
}

func benchmarkWaitGroupWait(b *testing.B, localWork int) {
	wg := newSyncWgCompatible()
	b.RunParallel(func(pb *testing.PB) {
		foo := 0
		for pb.Next() {
			wg.Wait()
			for i := 0; i < localWork; i++ {
				foo *= 2
				foo /= 2
			}
		}
		_ = foo
	})
}

func BenchmarkWaitGroupWait(b *testing.B) {
	benchmarkWaitGroupWait(b, 0)
}

func BenchmarkWaitGroupWaitWork(b *testing.B) {
	benchmarkWaitGroupWait(b, 100)
}

func BenchmarkWaitGroupActuallyWait(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			wg := newSyncWgCompatible()
			wg.Add(1)
			go func() {
				wg.Done()
			}()
			wg.Wait()
		}
	})
}
