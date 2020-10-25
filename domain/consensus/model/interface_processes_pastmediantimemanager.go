package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// PastMedianTimeManager provides a method to resolve the
// past median time of a block
type PastMedianTimeManager interface {
	PastMedianTime(blockHash *externalapi.DomainHash) (int64, error)
}
