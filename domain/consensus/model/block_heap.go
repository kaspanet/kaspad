package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// BlockHeap represents a heap of block hashes, providing a priority-queue functionality
type BlockHeap interface {
	Push(blockHash *externalapi.DomainHash) error
	Pop() *externalapi.DomainHash
	Len() int
}
