package model

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// BlockValidator exposes a set of validation classes, after which
// it's possible to determine whether a block is valid
type BlockValidator interface {
	ValidateHeaderInIsolation(stagingArea *StagingArea, blockHash *externalapi.DomainHash) error
	ValidateBodyInIsolation(stagingArea *StagingArea, blockHash *externalapi.DomainHash) error
	ValidateHeaderInContext(stagingArea *StagingArea, blockHash *externalapi.DomainHash, isBlockWithTrustedData bool) error
	ValidateBodyInContext(stagingArea *StagingArea, blockHash *externalapi.DomainHash, isBlockWithTrustedData bool) error
	ValidatePruningPointViolationAndProofOfWorkAndDifficulty(stagingArea *StagingArea, blockHash *externalapi.DomainHash, isBlockWithTrustedData bool) error
}
