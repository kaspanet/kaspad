package blockrelay

type ibdProgressReporter struct {
	lowDAAScore                 uint64
	highDAAScore                uint64
	objectName                  string
	totalDAAScoreDifference     uint64
	lastReportedProgressPercent int
	processed                   int
}

func newIBDProgressReporter(lowDAAScore uint64, highDAAScore uint64, objectName string) *ibdProgressReporter {
	return &ibdProgressReporter{
		lowDAAScore:                 lowDAAScore,
		highDAAScore:                highDAAScore,
		objectName:                  objectName,
		totalDAAScoreDifference:     highDAAScore - lowDAAScore,
		lastReportedProgressPercent: 0,
		processed:                   0,
	}
}

func (ipr *ibdProgressReporter) reportProgress(processedDelta int, highestProcessedDAAScore uint64) {
	ipr.processed += processedDelta

	relativeDAAScore := highestProcessedDAAScore - ipr.lowDAAScore
	progressPercent := int((float64(relativeDAAScore) / float64(ipr.totalDAAScoreDifference)) * 100)
	if progressPercent > ipr.lastReportedProgressPercent {
		log.Infof("IBD: Processed %d %s (%d%%)", ipr.processed, ipr.objectName, progressPercent)
		ipr.lastReportedProgressPercent = progressPercent
	}
}
