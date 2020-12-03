package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// BlockGHOSTDAGData represents GHOSTDAG data for some block
type BlockGHOSTDAGData interface {
	BlueScore() uint64
	SelectedParent() *externalapi.DomainHash
	MergeSetBlues() []*externalapi.DomainHash
	MergeSetReds() []*externalapi.DomainHash
	BluesAnticoneSizes() map[externalapi.DomainHash]KType
}

// KType defines the size of GHOSTDAG consensus algorithm K parameter.
type KType byte
