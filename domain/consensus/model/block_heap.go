package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

type BlockHeap interface {
	Push(blockHash *externalapi.DomainHash) error
	Pop() *externalapi.DomainHash
}
