package mstime

import "time"

const (
	nanosecondsInMillisecond = int64(time.Millisecond / time.Nanosecond)
	millisecondsInSecond     = int64(time.Second / time.Millisecond)
)

func Now() time.Time {
	return ReduceToMillisecondPrecision(time.Now())
}

func UnixMilliToTime(ms int64) time.Time {
	seconds := ms / millisecondsInSecond
	nanoseconds := (ms - seconds*millisecondsInSecond) * nanosecondsInMillisecond
	return time.Unix(ms/millisecondsInSecond, nanoseconds)
}

func TimeToUnixMilli(t time.Time) int64 {
	return t.UnixNano() / nanosecondsInMillisecond
}

func ReduceToMillisecondPrecision(t time.Time) time.Time {
	nanoseconds := int64(t.Nanosecond())
	millisecondPrecisionNanoSeconds := (nanoseconds / nanosecondsInMillisecond) * nanosecondsInMillisecond
	return time.Unix(t.Unix(), millisecondPrecisionNanoSeconds)
}
