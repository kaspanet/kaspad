package flowcontext

import (
	peerpkg "github.com/kaspanet/kaspad/app/protocol/peer"
	"github.com/kaspanet/kaspad/app/protocol/protocolerrors"
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
func (f *FlowContext) OnNewBlock(block *externalapi.DomainBlock,
	blockInsertionResult *externalapi.BlockInsertionResult) error {

	hash := consensushashing.BlockHash(block)
	log.Debugf("OnNewBlock start for block %s", hash)
	defer log.Debugf("OnNewBlock end for block %s", hash)

	unorphaningResults, err := f.UnorphanBlocks(block)
	if err != nil {
		return err
	}

	log.Debugf("OnNewBlock: block %s unorphaned %d blocks", hash, len(unorphaningResults))

	newBlocks := []*externalapi.DomainBlock{block}
	newBlockInsertionResults := []*externalapi.BlockInsertionResult{blockInsertionResult}
	for _, unorphaningResult := range unorphaningResults {
		newBlocks = append(newBlocks, unorphaningResult.block)
		newBlockInsertionResults = append(newBlockInsertionResults, unorphaningResult.blockInsertionResult)
	}

	allAcceptedTransactions := make([]*externalapi.DomainTransaction, 0)
	for i, newBlock := range newBlocks {
		log.Debugf("OnNewBlock: passing block %s transactions to mining manager", hash)
		acceptedTransactions, err := f.Domain().MiningManager().HandleNewBlockTransactions(newBlock.Transactions)
		if err != nil {
			return err
		}
		allAcceptedTransactions = append(allAcceptedTransactions, acceptedTransactions...)

		if f.onBlockAddedToDAGHandler != nil {
			log.Debugf("OnNewBlock: calling f.onBlockAddedToDAGHandler for block %s", hash)
			blockInsertionResult = newBlockInsertionResults[i]
			err := f.onBlockAddedToDAGHandler(newBlock, blockInsertionResult)
			if err != nil {
				return err
			}
		}
	}

	return f.broadcastTransactionsAfterBlockAdded(newBlocks, allAcceptedTransactions)
}

// OnPruningPointUTXOSetOverride calls the handler function whenever the UTXO set
// resets due to pruning point change via IBD.
func (f *FlowContext) OnPruningPointUTXOSetOverride() error {
	if f.onPruningPointUTXOSetOverrideHandler != nil {
		return f.onPruningPointUTXOSetOverrideHandler()
	}
	return nil
}

func (f *FlowContext) broadcastTransactionsAfterBlockAdded(
	addedBlocks []*externalapi.DomainBlock, transactionsAcceptedToMempool []*externalapi.DomainTransaction) error {

	f.updateTransactionsToRebroadcast(addedBlocks)

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
	if len(block.Transactions) == 0 {
		return protocolerrors.Errorf(false, "cannot add header only block")
	}

	blockInsertionResult, err := f.Domain().Consensus().ValidateAndInsertBlock(block)
	if err != nil {
		if errors.As(err, &ruleerrors.RuleError{}) {
			log.Warnf("Validation failed for block %s: %s", consensushashing.BlockHash(block), err)
		}
		return err
	}
	err = f.OnNewBlock(block, blockInsertionResult)
	if err != nil {
		return err
	}
	return f.Broadcast(appmessage.NewMsgInvBlock(consensushashing.BlockHash(block)))
}

// IsIBDRunning returns true if IBD is currently marked as running
func (f *FlowContext) IsIBDRunning() bool {
	f.ibdPeerMutex.RLock()
	defer f.ibdPeerMutex.RUnlock()

	return f.ibdPeer != nil
}

// TrySetIBDRunning attempts to set `isInIBD`. Returns false
// if it is already set
func (f *FlowContext) TrySetIBDRunning(ibdPeer *peerpkg.Peer) bool {
	f.ibdPeerMutex.Lock()
	defer f.ibdPeerMutex.Unlock()

	if f.ibdPeer != nil {
		return false
	}
	f.ibdPeer = ibdPeer
	log.Infof("IBD started")

	return true
}

// UnsetIBDRunning unsets isInIBD
func (f *FlowContext) UnsetIBDRunning() {
	f.ibdPeerMutex.Lock()
	defer f.ibdPeerMutex.Unlock()

	if f.ibdPeer == nil {
		panic("attempted to unset isInIBD when it was not set to begin with")
	}

	f.ibdPeer = nil
}

// IBDPeer returns the current IBD peer or null if the node is not
// in IBD
func (f *FlowContext) IBDPeer() *peerpkg.Peer {
	f.ibdPeerMutex.RLock()
	defer f.ibdPeerMutex.RUnlock()

	return f.ibdPeer
}
