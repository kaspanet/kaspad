package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// ConsensusStateStore represents a store for the current consensus state
type ConsensusStateStore interface {
	Stage(consensusStateChanges *ConsensusStateChanges)
	IsStaged() bool
	Discard()
	Commit(dbTx DBTxProxy) error
	UTXOByOutpoint(dbContext DBContextProxy, outpoint *externalapi.DomainOutpoint) (*externalapi.UTXOEntry, error)
}
