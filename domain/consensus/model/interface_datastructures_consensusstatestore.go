package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// ConsensusStateStore represents a store for the current consensus state
type ConsensusStateStore interface {
	Store
	Stage(consensusStateChanges *ConsensusStateChanges)
	IsStaged() bool
	UTXOByOutpoint(dbContext DBReader, outpoint *externalapi.DomainOutpoint) (*externalapi.UTXOEntry, error)

	Tips(dbContext DBReader) ([]*externalapi.DomainHash, error)
	SetTips(tipHashes []*externalapi.DomainHash) error
}
