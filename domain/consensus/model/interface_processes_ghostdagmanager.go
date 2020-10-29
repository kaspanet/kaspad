package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// GHOSTDAGManager resolves and manages GHOSTDAG block data
type GHOSTDAGManager interface {
	GHOSTDAG(blockHash *externalapi.DomainHash) error
	ChooseSelectedParent(blockHashA *externalapi.DomainHash, blockHashB *externalapi.DomainHash) (*externalapi.DomainHash, error)
	Less(blockHashA *externalapi.DomainHash, ghostdagDataA *BlockGHOSTDAGData,
		blockHashB *externalapi.DomainHash, ghostdagDataB *BlockGHOSTDAGData) bool
}
