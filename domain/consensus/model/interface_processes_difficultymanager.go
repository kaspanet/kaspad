package model

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// DifficultyManager provides a method to resolve the
// difficulty value of a block
type DifficultyManager interface {
	StageDAADataAndReturnRequiredDifficulty(stagingArea *StagingArea, blockHash *externalapi.DomainHash) (uint32, error)
	RequiredDifficulty(stagingArea *StagingArea, blockHash *externalapi.DomainHash) (uint32, error)
	EstimateNetworkHashesPerSecond(startHash *externalapi.DomainHash, windowSize int) (uint64, error)
}
