package model

import (
	"github.com/kaspanet/kaspad/app/appmessage"
)

// ConsensusStateStore represents a store for the current consensus state
type ConsensusStateStore interface {
	Update(dbTx TxContextProxy, utxoDiff *UTXODiff)
	UTXOByOutpoint(dbContext ContextProxy, outpoint *appmessage.Outpoint) *UTXOEntry
}
