package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// BlockIterator is an iterator over blocks according to some order.
type BlockIterator interface {
	Next() bool
	Get() *externalapi.DomainHash
}
