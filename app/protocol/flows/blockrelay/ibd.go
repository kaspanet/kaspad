package blockrelay

import (
	"time"

	"github.com/kaspanet/kaspad/infrastructure/logger"

	"github.com/kaspanet/kaspad/domain/consensus/model"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/protocol/common"
	"github.com/kaspanet/kaspad/app/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/pkg/errors"
)

func (flow *handleRelayInvsFlow) runIBDIfNotRunning(block *externalapi.DomainBlock) error {
	wasIBDNotRunning := flow.TrySetIBDRunning(flow.peer)
	if !wasIBDNotRunning {
		log.Debugf("IBD is already running")
		return nil
	}

	isFinishedSuccessfully := false
	defer func() {
		flow.UnsetIBDRunning()
		flow.logIBDFinished(isFinishedSuccessfully)
	}()

	highHash := consensushashing.BlockHash(block)
	log.Debugf("IBD started with peer %s and highHash %s", flow.peer, highHash)
	log.Debugf("Syncing blocks up to %s", highHash)
	log.Debugf("Trying to find highest shared chain block with peer %s with high hash %s", flow.peer, highHash)
	highestSharedBlockHash, highestSharedBlockFound, err := flow.findHighestSharedBlockHash(highHash)
	if err != nil {
		return err
	}
	log.Debugf("Found highest shared chain block %s with peer %s", highestSharedBlockHash, flow.peer)

	shouldDownloadHeadersProof, shouldSync, err := flow.shouldSyncAndShouldDownloadHeadersProof(block, highestSharedBlockFound)
	if err != nil {
		return err
	}

	if !shouldSync {
		return nil
	}

	if shouldDownloadHeadersProof {
		log.Infof("Starting IBD with headers proof")
		err := flow.ibdWithHeadersProof(highHash)
		if err != nil {
			return err
		}
	} else {
		if flow.Config().NetParams().DisallowDirectBlocksOnTopOfGenesis && !flow.Config().AllowSubmitBlockWhenNotSynced {
			isGenesisVirtualSelectedParent, err := flow.isGenesisVirtualSelectedParent()
			if err != nil {
				return err
			}

			if isGenesisVirtualSelectedParent {
				log.Infof("Cannot IBD to %s because it won't change the pruning point. The node needs to IBD "+
					"to the recent pruning point before normal operation can resume.", highHash)
				return nil
			}
		}

		err = flow.syncPruningPointFutureHeaders(flow.Domain().Consensus(), highestSharedBlockHash, highHash)
		if err != nil {
			return err
		}
	}

	err = flow.syncMissingBlockBodies(highHash)
	if err != nil {
		return err
	}

	log.Debugf("Finished syncing blocks up to %s", highHash)
	isFinishedSuccessfully = true
	return nil
}

func (flow *handleRelayInvsFlow) logIBDFinished(isFinishedSuccessfully bool) {
	successString := "successfully"
	if !isFinishedSuccessfully {
		successString = "(interrupted)"
	}
	log.Infof("IBD finished %s", successString)
}

// findHighestSharedBlock attempts to find the highest shared block between the peer
// and this node. This method may fail because the peer and us have conflicting pruning
// points. In that case we return (nil, false, nil) so that we may stop IBD gracefully.
func (flow *handleRelayInvsFlow) findHighestSharedBlockHash(
	targetHash *externalapi.DomainHash) (*externalapi.DomainHash, bool, error) {

	log.Debugf("Sending a blockLocator to %s between pruning point and headers selected tip", flow.peer)
	blockLocator, err := flow.Domain().Consensus().CreateFullHeadersSelectedChainBlockLocator()
	if err != nil {
		return nil, false, err
	}

	for {
		highestHash, highestHashFound, err := flow.fetchHighestHash(targetHash, blockLocator)
		if err != nil {
			return nil, false, err
		}
		if !highestHashFound {
			return nil, false, nil
		}
		highestHashIndex, err := flow.findHighestHashIndex(highestHash, blockLocator)
		if err != nil {
			return nil, false, err
		}

		if highestHashIndex == 0 ||
			// If the block locator contains only two adjacent chain blocks, the
			// syncer will always find the same highest chain block, so to avoid
			// an endless loop, we explicitly stop the loop in such situation.
			(len(blockLocator) == 2 && highestHashIndex == 1) {

			return highestHash, true, nil
		}

		locatorHashAboveHighestHash := highestHash
		if highestHashIndex > 0 {
			locatorHashAboveHighestHash = blockLocator[highestHashIndex-1]
		}

		blockLocator, err = flow.nextBlockLocator(highestHash, locatorHashAboveHighestHash)
		if err != nil {
			return nil, false, err
		}
	}
}

func (flow *handleRelayInvsFlow) nextBlockLocator(lowHash, highHash *externalapi.DomainHash) (externalapi.BlockLocator, error) {
	log.Debugf("Sending a blockLocator to %s between %s and %s", flow.peer, lowHash, highHash)
	blockLocator, err := flow.Domain().Consensus().CreateHeadersSelectedChainBlockLocator(lowHash, highHash)
	if err != nil {
		if errors.Is(model.ErrBlockNotInSelectedParentChain, err) {
			return nil, err
		}
		log.Debugf("Headers selected parent chain moved since findHighestSharedBlockHash - " +
			"restarting with full block locator")
		blockLocator, err = flow.Domain().Consensus().CreateFullHeadersSelectedChainBlockLocator()
		if err != nil {
			return nil, err
		}
	}

	return blockLocator, nil
}

func (flow *handleRelayInvsFlow) findHighestHashIndex(
	highestHash *externalapi.DomainHash, blockLocator externalapi.BlockLocator) (int, error) {

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
		return 0, protocolerrors.Errorf(true, "highest hash %s "+
			"returned from peer %s is not in the original blockLocator", highestHash, flow.peer)
	}
	log.Debugf("The index of the highest hash in the original "+
		"blockLocator sent to %s is %d", flow.peer, highestHashIndex)

	return highestHashIndex, nil
}

// fetchHighestHash attempts to fetch the highest hash the peer knows amongst the given
// blockLocator. This method may fail because the peer and us have conflicting pruning
// points. In that case we return (nil, false, nil) so that we may stop IBD gracefully.
func (flow *handleRelayInvsFlow) fetchHighestHash(
	targetHash *externalapi.DomainHash, blockLocator externalapi.BlockLocator) (*externalapi.DomainHash, bool, error) {

	ibdBlockLocatorMessage := appmessage.NewMsgIBDBlockLocator(targetHash, blockLocator)
	err := flow.outgoingRoute.Enqueue(ibdBlockLocatorMessage)
	if err != nil {
		return nil, false, err
	}
	message, err := flow.dequeueIncomingMessageAndSkipInvs(common.DefaultTimeout)
	if err != nil {
		return nil, false, err
	}
	switch message := message.(type) {
	case *appmessage.MsgIBDBlockLocatorHighestHash:
		highestHash := message.HighestHash
		log.Debugf("The highest hash the peer %s knows is %s", flow.peer, highestHash)

		return highestHash, true, nil
	case *appmessage.MsgIBDBlockLocatorHighestHashNotFound:
		log.Debugf("Peer %s does not know any block within our blockLocator. "+
			"This should only happen if there's a DAG split deeper than the pruning point.", flow.peer)
		return nil, false, nil
	default:
		return nil, false, protocolerrors.Errorf(true, "received unexpected message type. "+
			"expected: %s, got: %s", appmessage.CmdIBDBlockLocatorHighestHash, message.Command())
	}
}

func (flow *handleRelayInvsFlow) syncPruningPointFutureHeaders(consensus externalapi.Consensus, highestSharedBlockHash *externalapi.DomainHash,
	highHash *externalapi.DomainHash) error {

	log.Infof("Downloading headers from %s", flow.peer)

	err := flow.sendRequestHeaders(highestSharedBlockHash, highHash)
	if err != nil {
		return err
	}

	// Keep a short queue of BlockHeadersMessages so that there's
	// never a moment when the node is not validating and inserting
	// headers
	blockHeadersMessageChan := make(chan *appmessage.BlockHeadersMessage, 2)
	errChan := make(chan error)
	spawn("handleRelayInvsFlow-syncPruningPointFutureHeaders", func() {
		for {
			blockHeadersMessage, doneIBD, err := flow.receiveHeaders()
			if err != nil {
				errChan <- err
				return
			}
			if doneIBD {
				close(blockHeadersMessageChan)
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
		case ibdBlocksMessage, ok := <-blockHeadersMessageChan:
			if !ok {
				// If the highHash has not been received, the peer is misbehaving
				highHashBlockInfo, err := consensus.GetBlockInfo(highHash)
				if err != nil {
					return err
				}
				if !highHashBlockInfo.Exists {
					return protocolerrors.Errorf(true, "did not receive "+
						"highHash block %s from peer %s during block download", highHash, flow.peer)
				}
				return nil
			}
			for _, header := range ibdBlocksMessage.BlockHeaders {
				err = flow.processHeader(consensus, header)
				if err != nil {
					return err
				}
			}
		case err := <-errChan:
			return err
		}
	}
}

func (flow *handleRelayInvsFlow) sendRequestHeaders(highestSharedBlockHash *externalapi.DomainHash,
	peerSelectedTipHash *externalapi.DomainHash) error {

	msgGetBlockInvs := appmessage.NewMsgRequstHeaders(highestSharedBlockHash, peerSelectedTipHash)
	return flow.outgoingRoute.Enqueue(msgGetBlockInvs)
}

func (flow *handleRelayInvsFlow) receiveHeaders() (msgIBDBlock *appmessage.BlockHeadersMessage, doneHeaders bool, err error) {
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
				"expected: %s or %s, got: %s",
				appmessage.CmdBlockHeaders,
				appmessage.CmdDoneHeaders,
				message.Command())
	}
}

func (flow *handleRelayInvsFlow) processHeader(consensus externalapi.Consensus, msgBlockHeader *appmessage.MsgBlockHeader) error {
	header := appmessage.BlockHeaderToDomainBlockHeader(msgBlockHeader)
	block := &externalapi.DomainBlock{
		Header:       header,
		Transactions: nil,
	}

	blockHash := consensushashing.BlockHash(block)
	blockInfo, err := consensus.GetBlockInfo(blockHash)
	if err != nil {
		return err
	}
	if blockInfo.Exists {
		log.Debugf("Block header %s is already in the DAG. Skipping...", blockHash)
		return nil
	}
	_, err = consensus.ValidateAndInsertBlock(block, false)
	if err != nil {
		if !errors.As(err, &ruleerrors.RuleError{}) {
			return errors.Wrapf(err, "failed to process header %s during IBD", blockHash)
		}

		if errors.Is(err, ruleerrors.ErrDuplicateBlock) {
			log.Debugf("Skipping block header %s as it is a duplicate", blockHash)
		} else {
			log.Infof("Rejected block header %s from %s during IBD: %s", blockHash, flow.peer, err)
			return protocolerrors.Wrapf(true, err, "got invalid block header %s during IBD", blockHash)
		}
	}

	return nil
}

func (flow *handleRelayInvsFlow) validatePruningPointFutureHeaderTimestamps() error {
	headerSelectedTipHash, err := flow.Domain().StagingConsensus().GetHeadersSelectedTip()
	if err != nil {
		return err
	}
	headerSelectedTipHeader, err := flow.Domain().StagingConsensus().GetBlockHeader(headerSelectedTipHash)
	if err != nil {
		return err
	}
	headerSelectedTipTimestamp := headerSelectedTipHeader.TimeInMilliseconds()

	currentSelectedTipHash, err := flow.Domain().Consensus().GetHeadersSelectedTip()
	if err != nil {
		return err
	}
	currentSelectedTipHeader, err := flow.Domain().Consensus().GetBlockHeader(currentSelectedTipHash)
	if err != nil {
		return err
	}
	currentSelectedTipTimestamp := currentSelectedTipHeader.TimeInMilliseconds()

	if headerSelectedTipTimestamp < currentSelectedTipTimestamp {
		return protocolerrors.Errorf(false, "the timestamp of the candidate selected "+
			"tip is smaller than the current selected tip")
	}

	minTimestampDifferenceInMilliseconds := (10 * time.Minute).Milliseconds()
	if headerSelectedTipTimestamp-currentSelectedTipTimestamp < minTimestampDifferenceInMilliseconds {
		return protocolerrors.Errorf(false, "difference between the timestamps of "+
			"the current pruning point and the candidate pruning point is too small. Aborting IBD...")
	}
	return nil
}

func (flow *handleRelayInvsFlow) receiveAndInsertPruningPointUTXOSet(
	consensus externalapi.Consensus, pruningPointHash *externalapi.DomainHash) (bool, error) {

	onEnd := logger.LogAndMeasureExecutionTime(log, "receiveAndInsertPruningPointUTXOSet")
	defer onEnd()

	receivedChunkCount := 0
	receivedUTXOCount := 0
	for {
		message, err := flow.dequeueIncomingMessageAndSkipInvs(common.DefaultTimeout)
		if err != nil {
			return false, err
		}

		switch message := message.(type) {
		case *appmessage.MsgPruningPointUTXOSetChunk:
			receivedUTXOCount += len(message.OutpointAndUTXOEntryPairs)
			domainOutpointAndUTXOEntryPairs :=
				appmessage.OutpointAndUTXOEntryPairsToDomainOutpointAndUTXOEntryPairs(message.OutpointAndUTXOEntryPairs)

			err := consensus.AppendImportedPruningPointUTXOs(domainOutpointAndUTXOEntryPairs)
			if err != nil {
				return false, err
			}

			receivedChunkCount++
			if receivedChunkCount%ibdBatchSize == 0 {
				log.Debugf("Received %d UTXO set chunks so far, totaling in %d UTXOs",
					receivedChunkCount, receivedUTXOCount)

				requestNextPruningPointUTXOSetChunkMessage := appmessage.NewMsgRequestNextPruningPointUTXOSetChunk()
				err := flow.outgoingRoute.Enqueue(requestNextPruningPointUTXOSetChunkMessage)
				if err != nil {
					return false, err
				}
			}

		case *appmessage.MsgDonePruningPointUTXOSetChunks:
			log.Infof("Finished receiving the UTXO set. Total UTXOs: %d", receivedUTXOCount)
			return true, nil

		case *appmessage.MsgUnexpectedPruningPoint:
			log.Infof("Could not receive the next UTXO chunk because the pruning point %s "+
				"is no longer the pruning point of peer %s", pruningPointHash, flow.peer)
			return false, nil

		default:
			return false, protocolerrors.Errorf(true, "received unexpected message type. "+
				"expected: %s or %s or %s, got: %s", appmessage.CmdPruningPointUTXOSetChunk,
				appmessage.CmdDonePruningPointUTXOSetChunks, appmessage.CmdUnexpectedPruningPoint, message.Command(),
			)
		}
	}
}

func (flow *handleRelayInvsFlow) syncMissingBlockBodies(highHash *externalapi.DomainHash) error {
	hashes, err := flow.Domain().Consensus().GetMissingBlockBodyHashes(highHash)
	if err != nil {
		return err
	}
	if len(hashes) == 0 {
		// Blocks can be inserted inside the DAG during IBD if those were requested before IBD started.
		// In rare cases, all the IBD blocks might be already inserted by the time we reach this point.
		// In these cases - GetMissingBlockBodyHashes would return an empty array.
		log.Debugf("No missing block body hashes found.")
		return nil
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

			err = flow.banIfBlockIsHeaderOnly(block)
			if err != nil {
				return err
			}

			blockInsertionResult, err := flow.Domain().Consensus().ValidateAndInsertBlock(block, false)
			if err != nil {
				if errors.Is(err, ruleerrors.ErrDuplicateBlock) {
					log.Debugf("Skipping IBD Block %s as it has already been added to the DAG", blockHash)
					continue
				}
				return protocolerrors.ConvertToBanningProtocolErrorIfRuleError(err, "invalid block %s", blockHash)
			}
			err = flow.OnNewBlock(block, blockInsertionResult)
			if err != nil {
				return err
			}
		}
	}

	return flow.Domain().Consensus().ResolveVirtual()
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
