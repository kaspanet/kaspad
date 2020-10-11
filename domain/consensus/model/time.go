package model

import "time"

const (
	nanosecondsInAMillisecond = int64(time.Millisecond / time.Nanosecond)
)

// DomainTime is domain representation of time.Time that guarantees that all
// of its methods will return times with millisecond precision.
type DomainTime struct {
	time time.Time
}

// UnixMilliseconds returns t as a Unix time, the number of milliseconds elapsed
// since January 1, 1970 UTC.
func (t DomainTime) UnixMilliseconds() int64 {
	return t.time.UnixNano() / nanosecondsInAMillisecond
}

// ToDomainTime converts t to DomainTime
func ToDomainTime(t time.Time) *DomainTime {
	return &DomainTime{
		time: t.Round(time.Millisecond),
	}
}
