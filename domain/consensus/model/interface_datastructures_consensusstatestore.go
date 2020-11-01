package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// ConsensusStateStore represents a store for the current consensus state
type ConsensusStateStore interface {
	Store
	IsStaged() bool

	StageVirtualUTXODiff(virtualUTXODiff *UTXODiff)
	UTXOByOutpoint(dbContext DBReader, outpoint *externalapi.DomainOutpoint) (*externalapi.UTXOEntry, error)
	HasUTXOByOutpoint(dbContext DBReader, outpoint *externalapi.DomainOutpoint) (bool, error)

	StageVirtualDiffParents(virtualDiffParents []*externalapi.DomainHash) error
	VirtualDiffParents(dbContext DBReader) ([]*externalapi.DomainHash, error)

	Tips(dbContext DBReader) ([]*externalapi.DomainHash, error)
	StageTips(tipHashes []*externalapi.DomainHash) error
}
