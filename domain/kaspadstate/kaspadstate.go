package kaspadstate

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/kaspadstate/algorithms/blockprocessor"
	"github.com/kaspanet/kaspad/domain/kaspadstate/algorithms/consensusstatemanager"
	"github.com/kaspanet/kaspad/domain/kaspadstate/model"
	domainmodel "github.com/kaspanet/kaspad/domain/model"
	"github.com/kaspanet/kaspad/util"
)

type KaspadState interface {
	BuildBlock(transactionSelector domainmodel.TransactionSelector) *appmessage.MsgBlock
	ValidateAndInsertBlock(block *appmessage.MsgBlock) error

	UTXOByOutpoint(outpoint *appmessage.Outpoint) *model.UTXOEntry
	ValidateTransaction(transaction *util.Tx, utxoEntries []*model.UTXOEntry) error
}

type kaspadState struct {
	blockProcessor        blockprocessor.BlockProcessor
	consensusStateManager consensusstatemanager.ConsensusStateManager
}

func (s *kaspadState) BuildBlock(transactionSelector domainmodel.TransactionSelector) *appmessage.MsgBlock {
	return s.blockProcessor.BuildBlock(transactionSelector)
}

func (s *kaspadState) ValidateAndInsertBlock(block *appmessage.MsgBlock) error {
	return s.blockProcessor.ValidateAndInsertBlock(block)
}

func (s *kaspadState) UTXOByOutpoint(outpoint *appmessage.Outpoint) *model.UTXOEntry {
	return s.consensusStateManager.UTXOByOutpoint(outpoint)
}

func (s *kaspadState) ValidateTransaction(transaction *util.Tx, utxoEntries []*model.UTXOEntry) error {
	return s.consensusStateManager.ValidateTransaction(transaction, utxoEntries)
}
