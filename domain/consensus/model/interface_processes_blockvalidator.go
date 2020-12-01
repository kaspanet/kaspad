package model

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// BlockValidator exposes a set of validation classes, after which
// it's possible to determine whether a block is valid
type BlockValidator interface {
	ValidateHeaderInIsolation(blockHash *externalapi.DomainHash) error
	ValidateBodyInIsolation(blockHash *externalapi.DomainHash) error
	ValidateHeaderInContext(blockHash *externalapi.DomainHash) error
	ValidateBodyInContext(blockHash *externalapi.DomainHash) error
	ValidatePruningPointViolationAndProofOfWorkAndDifficulty(blockHash *externalapi.DomainHash) error
}
