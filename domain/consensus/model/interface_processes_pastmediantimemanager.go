package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// PastMedianTimeManager provides a method to resolve the
// past median time of a block
type PastMedianTimeManager interface {
	PastMedianTime(stagingArea *StagingArea, blockHash *externalapi.DomainHash, isBlockWithPrefilledData bool) (int64, error)
}
