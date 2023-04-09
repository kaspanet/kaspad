package mstime

import (
	"github.com/pkg/errors"
	"time"
)

const (
	nanosecondsInMillisecond = int64(time.Millisecond / time.Nanosecond)
	millisecondsInSecond     = int64(time.Second / time.Millisecond)
)

// Time is a wrapper for time.Time that guarantees all of its methods will return a millisecond precisioned times.
type Time struct {
	time time.Time
}

// UnixMilliseconds returns t as a Unix time, the number of milliseconds elapsed
// since January 1, 1970 UTC.
func (t Time) UnixMilliseconds() int64 {
	return t.time.UnixNano() / nanosecondsInMillisecond
}

// UnixSeconds returns t as a Unix time, the number of seconds elapsed
// since January 1, 1970 UTC.
func (t Time) UnixSeconds() int64 {
	return t.time.Unix()
}

// String returns the time formatted using the format string
//
//	"2006-01-02 15:04:05.999999999 -0700 MST"
func (t Time) String() string {
	return t.time.String()
}

// Clock returns the hour, minute, and second within the day specified by t.
func (t Time) Clock() (hour, min, sec int) {
	return t.time.Clock()
}

// Millisecond returns the millisecond offset within the second specified by t,
// in the range [0, 999].
func (t Time) Millisecond() int {
	return t.time.Nanosecond() / int(nanosecondsInMillisecond)
}

// Date returns the year, month, and day in which t occurs.
func (t Time) Date() (year int, month time.Month, day int) {
	return t.time.Date()
}

// After reports whether the time instant t is after u.
func (t Time) After(u Time) bool {
	return t.time.After(u.time)
}

// Before reports whether the time instant t is before u.
func (t Time) Before(u Time) bool {
	return t.time.Before(u.time)
}

// Add returns the time t+d.
// It panics if d has a precision greater than one millisecond (the duration has a non zero microseconds part).
func (t Time) Add(d time.Duration) Time {
	validateDurationPrecision(d)
	return newMSTime(t.time.Add(d))
}

// Sub returns the duration t-u. If the result exceeds the maximum (or minimum)
// value that can be stored in a Duration, the maximum (or minimum) duration
// will be returned.
// To compute t-d for a duration d, use t.Add(-d).
func (t Time) Sub(u Time) time.Duration {
	return t.time.Sub(u.time)
}

// IsZero reports whether t represents the zero time instant,
// January 1, year 1, 00:00:00 UTC.
func (t Time) IsZero() bool {
	return t.time.IsZero()
}

// ToNativeTime converts t to time.Time
func (t Time) ToNativeTime() time.Time {
	return t.time
}

// Now returns the current local time, with precision of one millisecond.
func Now() Time {
	return ToMSTime(time.Now())
}

// UnixMilliseconds returns the local Time corresponding to the given Unix time,
// ms milliseconds since January 1, 1970 UTC.
func UnixMilliseconds(ms int64) Time {
	seconds := ms / millisecondsInSecond
	nanoseconds := (ms - seconds*millisecondsInSecond) * nanosecondsInMillisecond
	return newMSTime(time.Unix(ms/millisecondsInSecond, nanoseconds))
}

// Since returns the time elapsed since t.
// It is shorthand for Now().Sub(t).
func Since(t Time) time.Duration {
	return Now().Sub(t)
}

// ToMSTime converts t to Time.
// See Time for details.
func ToMSTime(t time.Time) Time {
	return newMSTime(t.Round(time.Millisecond))
}

func newMSTime(t time.Time) Time {
	return Time{time: t}
}

func validateDurationPrecision(d time.Duration) {
	if d.Nanoseconds()%nanosecondsInMillisecond != 0 {
		panic(errors.Errorf("duration %s has lower precision than millisecond", d))
	}
}
