package mstime

import (
	"testing"
	"time"
)

func TestToMSTime(t *testing.T) {
	nativeTime1 := time.Unix(100, 5e6+800)
	if wantNano, gotNano := int64(100e9+5e6), ToMSTime(nativeTime1).time.UnixNano(); gotNano != wantNano {
		t.Fatalf("expected UnixNano %d but got %d", wantNano, gotNano)
	}

	nativeTime2 := time.Unix(500, 8e6)
	if wantNano, gotNano := int64(500e9+8e6), ToMSTime(nativeTime2).time.UnixNano(); gotNano != wantNano {
		t.Fatalf("expected UnixNano %d but got %d", wantNano, gotNano)
	}
}

func TestNow(t *testing.T) {
	if Now().time.UnixNano()%1e6 != 0 {
		t.Fatalf("Now() has higher precision than one millisecond")
	}
}

func TestAdd(t *testing.T) {
	tests := []struct {
		unixMilli         int64
		duration          time.Duration
		expectsPanics     bool
		expectedUnixMilli int64
	}{
		{
			unixMilli:     100,
			duration:      time.Nanosecond,
			expectsPanics: true,
		},
		{
			unixMilli:         100,
			duration:          time.Second + time.Nanosecond,
			expectsPanics:     true,
			expectedUnixMilli: 1100,
		},
		{
			unixMilli:         100,
			duration:          time.Second,
			expectsPanics:     false,
			expectedUnixMilli: 1100,
		},
	}
	for i, test := range tests {
		func() {
			defer func() {
				r := recover()
				if test.expectsPanics && r == nil {
					t.Fatalf("test #%d didn't panic when it was expected to", i)
				}
				if !test.expectsPanics && r != nil {
					t.Fatalf("test #%d panicked when it was not expected to", i)
				}
			}()
			mtime := UnixMilliseconds(100).Add(test.duration)
			if mtime.UnixMilliseconds() != test.expectedUnixMilli {
				t.Fatalf("test #%d expected UnixMilliseconds to be %d but got %d", i, test.expectedUnixMilli, mtime.UnixMilliseconds())
			}
		}()
	}
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("Add didn't panic when ")
		}
	}()
	UnixMilliseconds(100).Add(time.Nanosecond)
}
