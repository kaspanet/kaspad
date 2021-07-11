package daa

import "time"

type averageDuration struct {
	average    float64
	count      uint64
	sampleSize uint64
}

func newAverageDuration(sampleSize uint64) *averageDuration {
	return &averageDuration{
		average:    0,
		count:      0,
		sampleSize: sampleSize,
	}
}

func (ad *averageDuration) add(duration time.Duration) {
	durationNanoseconds := float64(duration.Nanoseconds())

	ad.count++
	if ad.count > ad.sampleSize {
		ad.count = ad.sampleSize
	}

	if ad.count == 1 {
		ad.average = durationNanoseconds
		return
	}

	ad.average = ad.average + ((durationNanoseconds - ad.average) / float64(ad.count))
}

func (ad *averageDuration) toDuration() time.Duration {
	return time.Duration(ad.average)
}
