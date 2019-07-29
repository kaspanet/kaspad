// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package locks

import (
	"sync/atomic"
	"testing"
)

// All of the tests, except TestAddAfterWait, are copied
// from the native sync/waitgroup_test.go (with some
// minor changes), to check that the new waitGroup
// behaves the same, except enabling the use of add()
// concurrently with wait()

// syncWaitGroupCompatible is a type that was created in order to
// make the use of waitGroup similar to the native one, so it'll
// be more convenient to use the same tests from sync/waitgroup_test.go
type syncWaitGroupCompatible struct {
	*waitGroup
}

func (swg *syncWaitGroupCompatible) add(delta int) {
	for i := 0; i < delta; i++ {
		swg.waitGroup.add()
	}
}

func newSyncWgCompatible() *syncWaitGroupCompatible {
	return &syncWaitGroupCompatible{
		waitGroup: newWaitGroup(),
	}
}

func testWaitGroup(t *testing.T, wg1 *syncWaitGroupCompatible, wg2 *syncWaitGroupCompatible) {
	n := 16
	wg1.add(n)
	wg2.add(n)
	exited := make(chan struct{}, n)
	for i := 0; i != n; i++ {
		go func(i int) {
			wg1.done()
			wg2.wait()
			exited <- struct{}{}
		}(i)
	}
	wg1.wait()
	for i := 0; i != n; i++ {
		select {
		case <-exited:
			t.Fatal("waitGroup released group too soon")
		default:
		}
		wg2.done()
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
	wg.add(1)
	wg.done()
	wg.done()
	t.Fatal("Should panic, because wg.counter should be negative (-1), which is not allowed")
}

func TestAddAfterWait(t *testing.T) {
	wg := newSyncWgCompatible()
	wg.add(1)
	syncChan := make(chan struct{})
	go func() {
		syncChan <- struct{}{}
		wg.wait()
		syncChan <- struct{}{}
	}()
	<-syncChan
	wg.add(1)
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
		wg.add(1)
		go func() {
			atomic.AddInt32(n, 1)
			wg.done()
		}()
		// spawn goroutine 2
		wg.add(1)
		go func() {
			atomic.AddInt32(n, 1)
			wg.done()
		}()
		// Wait for goroutine 1 and 2
		wg.wait()
		if atomic.LoadInt32(n) != 2 {
			t.Fatal("Spurious wakeup from Wait")
		}
	}
}

func TestWaitGroupAlign(t *testing.T) {
	type X struct {
		x  byte
		wg *syncWaitGroupCompatible
	}
	x := X{wg: newSyncWgCompatible()}
	x.wg.add(1)
	go func(x *X) {
		x.wg.done()
	}(&x)
	x.wg.wait()
}

func BenchmarkWaitGroupUncontended(b *testing.B) {
	type PaddedWaitGroup struct {
		*syncWaitGroupCompatible
		pad [128]uint8
	}
	b.RunParallel(func(pb *testing.PB) {
		wg := PaddedWaitGroup{
			syncWaitGroupCompatible: newSyncWgCompatible(),
		}
		for pb.Next() {
			wg.add(1)
			wg.done()
			wg.wait()
		}
	})
}

func benchmarkWaitGroupAdddone(b *testing.B, localWork int) {
	wg := newSyncWgCompatible()
	b.RunParallel(func(pb *testing.PB) {
		foo := 0
		for pb.Next() {
			wg.add(1)
			for i := 0; i < localWork; i++ {
				foo *= 2
				foo /= 2
			}
			wg.done()
		}
		_ = foo
	})
}

func BenchmarkWaitGroupAdddone(b *testing.B) {
	benchmarkWaitGroupAdddone(b, 0)
}

func BenchmarkWaitGroupAddDoneWork(b *testing.B) {
	benchmarkWaitGroupAdddone(b, 100)
}

func benchmarkWaitGroupwait(b *testing.B, localWork int) {
	wg := newSyncWgCompatible()
	b.RunParallel(func(pb *testing.PB) {
		foo := 0
		for pb.Next() {
			wg.wait()
			for i := 0; i < localWork; i++ {
				foo *= 2
				foo /= 2
			}
		}
		_ = foo
	})
}

func BenchmarkWaitGroupwait(b *testing.B) {
	benchmarkWaitGroupwait(b, 0)
}

func BenchmarkWaitGroupWaitWork(b *testing.B) {
	benchmarkWaitGroupwait(b, 100)
}

func BenchmarkWaitGroupActuallywait(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			wg := newSyncWgCompatible()
			wg.add(1)
			go func() {
				wg.done()
			}()
			wg.wait()
		}
	})
}
