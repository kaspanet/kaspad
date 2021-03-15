package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// DifficultyManager provides a method to resolve the
// difficulty value of a block
type DifficultyManager interface {
	StageDAADataAndReturnRequiredDifficulty(blockHash *externalapi.DomainHash) (uint32, error)
	RequiredDifficulty(blockHash *externalapi.DomainHash) (uint32, error)
}
