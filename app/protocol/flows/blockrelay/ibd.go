package blockrelay

import (
	"time"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/protocol/common"
	"github.com/kaspanet/kaspad/app/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/pkg/errors"
)

func (flow *handleRelayInvsFlow) runIBDIfNotRunning(highHash *externalapi.DomainHash) error {
	wasIBDNotRunning := flow.TrySetIBDRunning(flow.peer)
	if !wasIBDNotRunning {
		log.Debugf("IBD is already running")
		return nil
	}
	defer flow.UnsetIBDRunning()

	log.Debugf("IBD started with peer %s and highHash %s", flow.peer, highHash)

	log.Debugf("Syncing headers up to %s", highHash)
	err := flow.syncHeaders(highHash)
	if err != nil {
		return err
	}
	log.Debugf("Finished syncing headers up to %s", highHash)

	log.Debugf("Syncing the current pruning point UTXO set")
	syncedPruningPointUTXOSetSuccessfully, err := flow.syncPruningPointUTXOSet()
	if err != nil {
		return err
	}
	if !syncedPruningPointUTXOSetSuccessfully {
		log.Debugf("Aborting IBD because the pruning point UTXO set failed to sync")
		return nil
	}
	log.Debugf("Finished syncing the current pruning point UTXO set")

	log.Debugf("Downloading block bodies up to %s", highHash)
	err = flow.syncMissingBlockBodies(highHash)
	if err != nil {
		return err
	}
	log.Debugf("Finished downloading block bodies up to %s", highHash)

	return nil
}

func (flow *handleRelayInvsFlow) syncHeaders(highHash *externalapi.DomainHash) error {
	highHashReceived := false
	for !highHashReceived {
		log.Debugf("Trying to find highest shared chain block with peer %s with high hash %s", flow.peer, highHash)
		highestSharedBlockHash, err := flow.findHighestSharedBlockHash(highHash)
		if err != nil {
			return err
		}
		log.Debugf("Found highest shared chain block %s with peer %s", highestSharedBlockHash, flow.peer)

		err = flow.downloadHeaders(highestSharedBlockHash, highHash)
		if err != nil {
			return err
		}

		// We're finished once highHash has been inserted into the DAG
		blockInfo, err := flow.Domain().Consensus().GetBlockInfo(highHash)
		if err != nil {
			return err
		}
		highHashReceived = blockInfo.Exists
		log.Debugf("Headers downloaded from peer %s. Are further headers required: %t", flow.peer, !highHashReceived)
	}
	return nil
}

func (flow *handleRelayInvsFlow) findHighestSharedBlockHash(targetHash *externalapi.DomainHash) (*externalapi.DomainHash, error) {
	lowHash, err := flow.Domain().Consensus().PruningPoint()
	if err != nil {
		return nil, err
	}
	highHash, err := flow.Domain().Consensus().GetHeadersSelectedTip()
	if err != nil {
		return nil, err
	}

	for !lowHash.Equal(highHash) {
		log.Debugf("Sending a blockLocator to %s between %s and %s", flow.peer, lowHash, highHash)
		blockLocator, err := flow.Domain().Consensus().CreateHeadersSelectedChainBlockLocator(lowHash, highHash)
		if err != nil {
			return nil, err
		}

		ibdBlockLocatorMessage := appmessage.NewMsgIBDBlockLocator(targetHash, blockLocator)
		err = flow.outgoingRoute.Enqueue(ibdBlockLocatorMessage)
		if err != nil {
			return nil, err
		}
		message, err := flow.dequeueIncomingMessageAndSkipInvs(common.DefaultTimeout)
		if err != nil {
			return nil, err
		}
		ibdBlockLocatorHighestHashMessage, ok := message.(*appmessage.MsgIBDBlockLocatorHighestHash)
		if !ok {
			return nil, protocolerrors.Errorf(true, "received unexpected message type. "+
				"expected: %s, got: %s", appmessage.CmdIBDBlockLocatorHighestHash, message.Command())
		}
		highestHash := ibdBlockLocatorHighestHashMessage.HighestHash
		log.Debugf("The highest hash the peer %s knows is %s", flow.peer, highestHash)

		highestHashIndex := 0
		highestHashIndexFound := false
		for i, blockLocatorHash := range blockLocator {
			if highestHash.Equal(blockLocatorHash) {
				highestHashIndex = i
				highestHashIndexFound = true
				break
			}
		}
		if !highestHashIndexFound {
			return nil, protocolerrors.Errorf(true, "highest hash %s "+
				"returned from peer %s is not in the original blockLocator", highestHash, flow.peer)
		}
		log.Debugf("The index of the highest hash in the original "+
			"blockLocator sent to %s is %d", flow.peer, highestHashIndex)

		// If the block locator contains only two adjacent chain blocks, the
		// syncer will always find the same highest chain block, so to avoid
		// an endless loop, we explicitly stop the loop in such situation.
		if len(blockLocator) == 2 && highestHashIndex == 1 {
			return highestHash, nil
		}

		locatorHashAboveHighestHash := highestHash
		if highestHashIndex > 0 {
			locatorHashAboveHighestHash = blockLocator[highestHashIndex-1]
		}
		highHash = locatorHashAboveHighestHash
		lowHash = highestHash
	}
	return highHash, nil
}

func (flow *handleRelayInvsFlow) downloadHeaders(highestSharedBlockHash *externalapi.DomainHash,
	highHash *externalapi.DomainHash) error {

	err := flow.sendRequestHeaders(highestSharedBlockHash, highHash)
	if err != nil {
		return err
	}

	// Keep a short queue of blockHeadersMessages so that there's
	// never a moment when the node is not validating and inserting
	// headers
	blockHeadersMessageChan := make(chan *appmessage.BlockHeadersMessage, 2)
	errChan := make(chan error)
	doneChan := make(chan interface{})
	spawn("handleRelayInvsFlow-downloadHeaders", func() {
		for {
			blockHeadersMessage, doneIBD, err := flow.receiveHeaders()
			if err != nil {
				errChan <- err
				return
			}
			if doneIBD {
				doneChan <- struct{}{}
				return
			}

			blockHeadersMessageChan <- blockHeadersMessage

			err = flow.outgoingRoute.Enqueue(appmessage.NewMsgRequestNextHeaders())
			if err != nil {
				errChan <- err
				return
			}
		}
	})

	for {
		select {
		case blockHeadersMessage := <-blockHeadersMessageChan:
			for _, header := range blockHeadersMessage.BlockHeaders {
				err = flow.processHeader(header)
				if err != nil {
					return err
				}
			}
		case err := <-errChan:
			return err
		case <-doneChan:
			return nil
		}
	}
}

func (flow *handleRelayInvsFlow) sendRequestHeaders(highestSharedBlockHash *externalapi.DomainHash,
	peerSelectedTipHash *externalapi.DomainHash) error {

	msgGetBlockInvs := appmessage.NewMsgRequstHeaders(highestSharedBlockHash, peerSelectedTipHash)
	return flow.outgoingRoute.Enqueue(msgGetBlockInvs)
}

func (flow *handleRelayInvsFlow) receiveHeaders() (msgIBDBlock *appmessage.BlockHeadersMessage, doneIBD bool, err error) {
	message, err := flow.dequeueIncomingMessageAndSkipInvs(common.DefaultTimeout)
	if err != nil {
		return nil, false, err
	}
	switch message := message.(type) {
	case *appmessage.BlockHeadersMessage:
		return message, false, nil
	case *appmessage.MsgDoneHeaders:
		return nil, true, nil
	default:
		return nil, false,
			protocolerrors.Errorf(true, "received unexpected message type. "+
				"expected: %s or %s, got: %s", appmessage.CmdHeader, appmessage.CmdDoneHeaders, message.Command())
	}
}

func (flow *handleRelayInvsFlow) processHeader(msgBlockHeader *appmessage.MsgBlockHeader) error {
	header := appmessage.BlockHeaderToDomainBlockHeader(msgBlockHeader)
	block := &externalapi.DomainBlock{
		Header:       header,
		Transactions: nil,
	}

	blockHash := consensushashing.BlockHash(block)
	blockInfo, err := flow.Domain().Consensus().GetBlockInfo(blockHash)
	if err != nil {
		return err
	}
	if blockInfo.Exists {
		log.Debugf("Block header %s is already in the DAG. Skipping...", blockHash)
		return nil
	}
	_, err = flow.Domain().Consensus().ValidateAndInsertBlock(block)
	if err != nil {
		if !errors.As(err, &ruleerrors.RuleError{}) {
			return errors.Wrapf(err, "failed to process header %s during IBD", blockHash)
		}
		log.Infof("Rejected block header %s from %s during IBD: %s", blockHash, flow.peer, err)

		return protocolerrors.Wrapf(true, err, "got invalid block %s during IBD", blockHash)
	}

	return nil
}

func (flow *handleRelayInvsFlow) syncPruningPointUTXOSet() (bool, error) {
	log.Debugf("Checking if a new pruning point is available")
	err := flow.outgoingRoute.Enqueue(appmessage.NewMsgRequestIBDRootHashMessage())
	if err != nil {
		return false, err
	}
	message, err := flow.dequeueIncomingMessageAndSkipInvs(common.DefaultTimeout)
	if err != nil {
		return false, err
	}
	msgIBDRootHash, ok := message.(*appmessage.MsgIBDRootHashMessage)
	if !ok {
		return false, protocolerrors.Errorf(true, "received unexpected message type. "+
			"expected: %s, got: %s", appmessage.CmdIBDRootHash, message.Command())
	}

	blockInfo, err := flow.Domain().Consensus().GetBlockInfo(msgIBDRootHash.Hash)
	if err != nil {
		return false, err
	}

	if blockInfo.BlockStatus != externalapi.StatusHeaderOnly {
		log.Debugf("Already has the block data of the new suggested pruning point %s", msgIBDRootHash.Hash)
		return true, nil
	}

	log.Infof("Checking if the suggested pruning point %s is compatible to the node DAG", msgIBDRootHash.Hash)
	isValid, err := flow.Domain().Consensus().IsValidPruningPoint(msgIBDRootHash.Hash)
	if err != nil {
		return false, err
	}

	if !isValid {
		log.Infof("The suggested pruning point %s is incompatible to this node DAG, so stopping IBD with this"+
			" peer", msgIBDRootHash.Hash)
		return false, nil
	}

	log.Info("Fetching the pruning point UTXO set")
	succeed, err := flow.fetchMissingUTXOSet(msgIBDRootHash.Hash)
	if err != nil {
		return false, err
	}

	if !succeed {
		log.Infof("Couldn't successfully fetch the pruning point UTXO set. Stopping IBD.")
		return false, nil
	}

	log.Info("Fetched the new pruning point UTXO set")
	return true, nil
}

func (flow *handleRelayInvsFlow) fetchMissingUTXOSet(ibdRootHash *externalapi.DomainHash) (succeed bool, err error) {
	err = flow.outgoingRoute.Enqueue(appmessage.NewMsgRequestIBDRootUTXOSetAndBlock(ibdRootHash))
	if err != nil {
		return false, err
	}

	utxoSet, block, found, err := flow.receiveIBDRootUTXOSetAndBlock()
	if err != nil {
		return false, err
	}

	if !found {
		return false, nil
	}

	err = flow.Domain().Consensus().ValidateAndInsertPruningPoint(block, utxoSet)
	if err != nil {
		// TODO: Find a better way to deal with finality conflicts.
		if errors.Is(err, ruleerrors.ErrSuggestedPruningViolatesFinality) {
			return false, nil
		}
		return false, protocolerrors.ConvertToBanningProtocolErrorIfRuleError(err, "error with IBD root UTXO set")
	}

	return true, nil
}

func (flow *handleRelayInvsFlow) receiveIBDRootUTXOSetAndBlock() ([]byte, *externalapi.DomainBlock, bool, error) {
	message, err := flow.dequeueIncomingMessageAndSkipInvs(common.DefaultTimeout)
	if err != nil {
		return nil, nil, false, err
	}

	var block *externalapi.DomainBlock
	switch message := message.(type) {
	case *appmessage.MsgIBDBlock:
		block = appmessage.MsgBlockToDomainBlock(message.MsgBlock)
	case *appmessage.MsgIBDRootNotFound:
		return nil, nil, false, nil
	default:
		return nil, nil, false,
			protocolerrors.Errorf(true, "received unexpected message type. "+
				"expected: %s or %s, got: %s",
				appmessage.CmdIBDBlock, appmessage.CmdIBDRootNotFound, message.Command(),
			)
	}
	log.Debugf("Received IBD root block %s", consensushashing.BlockHash(block))

	serializedUTXOSet := []byte{}
	receivedAllChunks := false
	receivedChunkCount := 0
	for !receivedAllChunks {
		message, err := flow.dequeueIncomingMessageAndSkipInvs(common.DefaultTimeout)
		if err != nil {
			return nil, nil, false, err
		}

		switch message := message.(type) {
		case *appmessage.MsgIBDRootUTXOSetChunk:
			serializedUTXOSet = append(serializedUTXOSet, message.Chunk...)
		case *appmessage.MsgDoneIBDRootUTXOSetChunks:
			receivedAllChunks = true
		default:
			return nil, nil, false,
				protocolerrors.Errorf(true, "received unexpected message type. "+
					"expected: %s or %s, got: %s",
					appmessage.CmdIBDRootUTXOSetChunk, appmessage.CmdDoneIBDRootUTXOSetChunks, message.Command(),
				)
		}

		receivedChunkCount++
		if !receivedAllChunks && receivedChunkCount%ibdBatchSize == 0 {
			log.Debugf("Received %d UTXO set chunks so far, totaling in %d bytes",
				receivedChunkCount, len(serializedUTXOSet))

			requestNextIBDRootUTXOSetChunkMessage := appmessage.NewMsgRequestNextIBDRootUTXOSetChunk()
			err := flow.outgoingRoute.Enqueue(requestNextIBDRootUTXOSetChunkMessage)
			if err != nil {
				return nil, nil, false, err
			}
		}
	}
	log.Debugf("Finished receiving the UTXO set. Total bytes: %d", len(serializedUTXOSet))

	return serializedUTXOSet, block, true, nil
}

func (flow *handleRelayInvsFlow) syncMissingBlockBodies(highHash *externalapi.DomainHash) error {
	hashes, err := flow.Domain().Consensus().GetMissingBlockBodyHashes(highHash)
	if err != nil {
		return err
	}

	for offset := 0; offset < len(hashes); offset += ibdBatchSize {
		var hashesToRequest []*externalapi.DomainHash
		if offset+ibdBatchSize < len(hashes) {
			hashesToRequest = hashes[offset : offset+ibdBatchSize]
		} else {
			hashesToRequest = hashes[offset:]
		}

		err := flow.outgoingRoute.Enqueue(appmessage.NewMsgRequestIBDBlocks(hashesToRequest))
		if err != nil {
			return err
		}

		for _, expectedHash := range hashesToRequest {
			message, err := flow.dequeueIncomingMessageAndSkipInvs(common.DefaultTimeout)
			if err != nil {
				return err
			}

			msgIBDBlock, ok := message.(*appmessage.MsgIBDBlock)
			if !ok {
				return protocolerrors.Errorf(true, "received unexpected message type. "+
					"expected: %s, got: %s", appmessage.CmdIBDBlock, message.Command())
			}

			block := appmessage.MsgBlockToDomainBlock(msgIBDBlock.MsgBlock)
			blockHash := consensushashing.BlockHash(block)
			if !expectedHash.Equal(blockHash) {
				return protocolerrors.Errorf(true, "expected block %s but got %s", expectedHash, blockHash)
			}

			blockInsertionResult, err := flow.Domain().Consensus().ValidateAndInsertBlock(block)
			if err != nil {
				return protocolerrors.ConvertToBanningProtocolErrorIfRuleError(err, "invalid block %s", blockHash)
			}
			err = flow.OnNewBlock(block, blockInsertionResult)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// dequeueIncomingMessageAndSkipInvs is a convenience method to be used during
// IBD. Inv messages are expected to arrive at any given moment, but should be
// ignored while we're in IBD
func (flow *handleRelayInvsFlow) dequeueIncomingMessageAndSkipInvs(timeout time.Duration) (appmessage.Message, error) {
	for {
		message, err := flow.incomingRoute.DequeueWithTimeout(timeout)
		if err != nil {
			return nil, err
		}
		if _, ok := message.(*appmessage.MsgInvRelayBlock); !ok {
			return message, nil
		}
	}
}
