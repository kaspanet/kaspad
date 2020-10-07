package model

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/util"
)

// ConsensusStateManager manages the node's consensus state
type ConsensusStateManager interface {
	UTXOByOutpoint(outpoint *appmessage.Outpoint) *UTXOEntry
	ValidateTransaction(transaction *util.Tx, utxoEntries []*UTXOEntry) error
	CalculateConsensusStateChanges(block *appmessage.MsgBlock) *ConsensusStateChanges
}
