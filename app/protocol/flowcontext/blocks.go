package flowcontext

import (
	"sync/atomic"

	"github.com/kaspanet/kaspad/app/protocol/blocklogger"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/protocol/flows/blockrelay"
)

// OnNewBlock updates the mempool after a new block arrival, and
// relays newly unorphaned transactions and possibly rebroadcast
// manually added transactions when not in IBD.
func (f *FlowContext) OnNewBlock(block *externalapi.DomainBlock) error {
	hash := consensushashing.BlockHash(block)
	log.Debugf("OnNewBlock start for block %s", hash)
	defer log.Debugf("OnNewBlock end for block %s", hash)
	unorphanedBlocks, err := f.UnorphanBlocks(block)
	if err != nil {
		return err
	}

	log.Debugf("OnNewBlock: block %s unorphaned %d blocks", hash, len(unorphanedBlocks))

	newBlocks := append([]*externalapi.DomainBlock{block}, unorphanedBlocks...)
	for _, newBlock := range newBlocks {
		blocklogger.LogBlock(block)

		log.Tracef("OnNewBlock: passing block %s transactions to mining manager", hash)
		_ = f.Domain().MiningManager().HandleNewBlockTransactions(newBlock.Transactions)

		if f.onBlockAddedToDAGHandler != nil {
			log.Tracef("OnNewBlock: calling f.onBlockAddedToDAGHandler for block %s", hash)
			err := f.onBlockAddedToDAGHandler(newBlock)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (f *FlowContext) broadcastTransactionsAfterBlockAdded(
	block *externalapi.DomainBlock, transactionsAcceptedToMempool []*externalapi.DomainTransaction) error {

	f.updateTransactionsToRebroadcast(block)

	// Don't relay transactions when in IBD.
	if f.IsIBDRunning() {
		return nil
	}

	var txIDsToRebroadcast []*externalapi.DomainTransactionID
	if f.shouldRebroadcastTransactions() {
		txIDsToRebroadcast = f.txIDsToRebroadcast()
	}

	txIDsToBroadcast := make([]*externalapi.DomainTransactionID, len(transactionsAcceptedToMempool)+len(txIDsToRebroadcast))
	for i, tx := range transactionsAcceptedToMempool {
		txIDsToBroadcast[i] = consensushashing.TransactionID(tx)
	}
	offset := len(transactionsAcceptedToMempool)
	for i, txID := range txIDsToRebroadcast {
		txIDsToBroadcast[offset+i] = txID
	}

	if len(txIDsToBroadcast) == 0 {
		return nil
	}
	if len(txIDsToBroadcast) > appmessage.MaxInvPerTxInvMsg {
		txIDsToBroadcast = txIDsToBroadcast[:appmessage.MaxInvPerTxInvMsg]
	}
	inv := appmessage.NewMsgInvTransaction(txIDsToBroadcast)
	return f.Broadcast(inv)
}

// SharedRequestedBlocks returns a *blockrelay.SharedRequestedBlocks for sharing
// data about requested blocks between different peers.
func (f *FlowContext) SharedRequestedBlocks() *blockrelay.SharedRequestedBlocks {
	return f.sharedRequestedBlocks
}

// AddBlock adds the given block to the DAG and propagates it.
func (f *FlowContext) AddBlock(block *externalapi.DomainBlock) error {
	err := f.Domain().Consensus().ValidateAndInsertBlock(block)
	if err != nil {
		if errors.As(err, &ruleerrors.RuleError{}) {
			log.Infof("Validation failed for block %s: %s", consensushashing.BlockHash(block), err)
			return nil
		}
		return err
	}
	err = f.OnNewBlock(block)
	if err != nil {
		return err
	}
	return f.Broadcast(appmessage.NewMsgInvBlock(consensushashing.BlockHash(block)))
}

// IsIBDRunning returns true if IBD is currently marked as running
func (f *FlowContext) IsIBDRunning() bool {
	return atomic.LoadUint32(&f.isInIBD) != 0
}

// TrySetIBDRunning attempts to set `isInIBD`. Returns false
// if it is already set
func (f *FlowContext) TrySetIBDRunning() bool {
	succeeded := atomic.CompareAndSwapUint32(&f.isInIBD, 0, 1)
	if succeeded {
		log.Infof("IBD started")
	}
	return succeeded
}

// UnsetIBDRunning unsets isInIBD
func (f *FlowContext) UnsetIBDRunning() {
	succeeded := atomic.CompareAndSwapUint32(&f.isInIBD, 1, 0)
	if !succeeded {
		panic("attempted to unset isInIBD when it was not set to begin with")
	}
	log.Infof("IBD finished")
}
