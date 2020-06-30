package mstime

import (
	"github.com/pkg/errors"
	"time"
)

const (
	nanosecondsInMillisecond = int64(time.Millisecond / time.Nanosecond)
	millisecondsInSecond     = int64(time.Second / time.Millisecond)
)

type Time struct {
	time time.Time
}

func (t Time) UnixMilli() int64 {
	return t.time.UnixNano() / nanosecondsInMillisecond
}

func (t Time) UnixSeconds() int64 {
	return t.time.Unix()
}

func (t Time) String() string {
	return t.time.String()
}

func (t Time) Clock() (hour, min, sec int) {
	return t.time.Clock()
}

func (t Time) Millisecond() int {
	return t.time.Nanosecond() / int(nanosecondsInMillisecond)
}

func (t Time) Date() (year int, month time.Month, day int) {
	return t.time.Date()
}

func (t Time) After(u Time) bool {
	return t.time.After(u.time)
}

func (t Time) Before(u Time) bool {
	return t.time.Before(u.time)
}

func (t Time) Add(d time.Duration) Time {
	validateDurationPrecision(d)
	return newMSTime(t.time.Add(d))
}

func (t Time) Sub(u Time) time.Duration {
	return t.time.Sub(u.time)
}

func (t Time) IsZero() bool {
	return t.time.IsZero()
}

func (t Time) ToGoTime() time.Time {
	return t.time
}

func Now() Time {
	return ToMSTime(time.Now())
}

func UnixMilli(ms int64) Time {
	seconds := ms / millisecondsInSecond
	nanoseconds := (ms - seconds*millisecondsInSecond) * nanosecondsInMillisecond
	return newMSTime(time.Unix(ms/millisecondsInSecond, nanoseconds))
}

func Since(t Time) time.Duration {
	return Now().Sub(t)
}

func ToMSTime(t time.Time) Time {
	nanoseconds := int64(t.Nanosecond())
	millisecondPrecisionNanoSeconds := (nanoseconds / nanosecondsInMillisecond) * nanosecondsInMillisecond
	return newMSTime(time.Unix(t.Unix(), millisecondPrecisionNanoSeconds))
}

func newMSTime(t time.Time) Time {
	return Time{time: t}
}

func validateDurationPrecision(d time.Duration) {
	if d.Nanoseconds()%nanosecondsInMillisecond != 0 {
		panic(errors.Errorf("duration %s has lower precision than millisecond", d))
	}
}
