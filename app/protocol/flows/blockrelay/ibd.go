package blockrelay

import (
	"fmt"
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

func (flow *handleRelayInvsFlow) runIBDIfNotRunning(highHash *externalapi.DomainHash) error {
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

	log.Debugf("IBD started with peer %s and highHash %s", flow.peer, highHash)
	log.Debugf("Syncing headers up to %s", highHash)
	headersSynced, err := flow.syncHeaders(highHash)
	if err != nil {
		return err
	}
	if !headersSynced {
		log.Debugf("Aborting IBD because the headers failed to sync")
		return nil
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

	isFinishedSuccessfully = true

	return nil
}

func (flow *handleRelayInvsFlow) logIBDFinished(isFinishedSuccessfully bool) {
	successString := "successfully"
	if !isFinishedSuccessfully {
		successString = "ahead of time"
	}
	log.Infof("IBD finished %s", successString)
}

// syncHeaders attempts to sync headers from the peer. This method may fail
// because the peer and us have conflicting pruning points. In that case we
// return (false, nil) so that we may stop IBD gracefully.
func (flow *handleRelayInvsFlow) syncHeaders(highHash *externalapi.DomainHash) (bool, error) {
	log.Debugf("Trying to find highest shared chain block with peer %s with high hash %s", flow.peer, highHash)
	highestSharedBlockHash, highestSharedBlockFound, err := flow.findHighestSharedBlockHash(highHash)
	if err != nil {
		return false, err
	}
	if !highestSharedBlockFound {
		return false, nil
	}
	log.Debugf("Found highest shared chain block %s with peer %s", highestSharedBlockHash, flow.peer)

	err = flow.downloadHeaders(highestSharedBlockHash, highHash)
	if err != nil {
		return false, err
	}

	// If the highHash has not been received, the peer is misbehaving
	highHashBlockInfo, err := flow.Domain().Consensus().GetBlockInfo(highHash)
	if err != nil {
		return false, err
	}
	if !highHashBlockInfo.Exists {
		return false, protocolerrors.Errorf(true, "did not receive "+
			"highHash header %s from peer %s during header download", highHash, flow.peer)
	}
	log.Debugf("Headers downloaded from peer %s", flow.peer)
	return true, nil
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
	spawn("handleRelayInvsFlow-downloadHeaders", func() {
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
		case blockHeadersMessage, ok := <-blockHeadersMessageChan:
			if !ok {
				return nil
			}
			for _, header := range blockHeadersMessage.BlockHeaders {
				err = flow.processHeader(header)
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

		if errors.Is(err, ruleerrors.ErrDuplicateBlock) {
			log.Debugf("Skipping block header %s as it is a duplicate", blockHash)
		} else {
			log.Infof("Rejected block header %s from %s during IBD: %s", blockHash, flow.peer, err)
			return protocolerrors.Wrapf(true, err, "got invalid block header %s during IBD", blockHash)
		}
	}

	return nil
}

func (flow *handleRelayInvsFlow) syncPruningPointUTXOSet() (bool, error) {
	log.Debugf("Checking if a new pruning point is available")
	err := flow.outgoingRoute.Enqueue(appmessage.NewMsgRequestPruningPointHashMessage())
	if err != nil {
		return false, err
	}
	message, err := flow.dequeueIncomingMessageAndSkipInvs(common.DefaultTimeout)
	if err != nil {
		return false, err
	}
	msgPruningPointHash, ok := message.(*appmessage.MsgPruningPointHashMessage)
	if !ok {
		return false, protocolerrors.Errorf(true, "received unexpected message type. "+
			"expected: %s, got: %s", appmessage.CmdPruningPointHash, message.Command())
	}

	blockInfo, err := flow.Domain().Consensus().GetBlockInfo(msgPruningPointHash.Hash)
	if err != nil {
		return false, err
	}

	if !blockInfo.Exists {
		return false, errors.Errorf("The pruning point header is missing")
	}

	if blockInfo.BlockStatus != externalapi.StatusHeaderOnly {
		log.Debugf("Already has the block data of the new suggested pruning point %s", msgPruningPointHash.Hash)
		return true, nil
	}

	log.Infof("Checking if the suggested pruning point %s is compatible to the node DAG", msgPruningPointHash.Hash)
	isValid, err := flow.Domain().Consensus().IsValidPruningPoint(msgPruningPointHash.Hash)
	if err != nil {
		return false, err
	}

	if !isValid {
		log.Infof("The suggested pruning point %s is incompatible to this node DAG, so stopping IBD with this"+
			" peer", msgPruningPointHash.Hash)
		return false, nil
	}

	log.Info("Fetching the pruning point UTXO set")
	succeed, err := flow.fetchMissingUTXOSet(msgPruningPointHash.Hash)
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

func (flow *handleRelayInvsFlow) fetchMissingUTXOSet(pruningPointHash *externalapi.DomainHash) (succeed bool, err error) {
	defer func() {
		err := flow.Domain().Consensus().ClearImportedPruningPointData()
		if err != nil {
			panic(fmt.Sprintf("failed to clear imported pruning point data: %s", err))
		}
	}()

	err = flow.outgoingRoute.Enqueue(appmessage.NewMsgRequestPruningPointUTXOSetAndBlock(pruningPointHash))
	if err != nil {
		return false, err
	}

	block, err := flow.receivePruningPointBlock()
	if err != nil {
		return false, err
	}

	receivedAll, err := flow.receiveAndInsertPruningPointUTXOSet(pruningPointHash)
	if err != nil {
		return false, err
	}
	if !receivedAll {
		return false, nil
	}

	err = flow.Domain().Consensus().ValidateAndInsertImportedPruningPoint(block)
	if err != nil {
		// TODO: Find a better way to deal with finality conflicts.
		if errors.Is(err, ruleerrors.ErrSuggestedPruningViolatesFinality) {
			return false, nil
		}
		return false, protocolerrors.ConvertToBanningProtocolErrorIfRuleError(err, "error with pruning point UTXO set")
	}

	err = flow.OnPruningPointUTXOSetOverride()
	if err != nil {
		return false, err
	}

	return true, nil
}

func (flow *handleRelayInvsFlow) receivePruningPointBlock() (*externalapi.DomainBlock, error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "receivePruningPointBlock")
	defer onEnd()

	message, err := flow.dequeueIncomingMessageAndSkipInvs(common.DefaultTimeout)
	if err != nil {
		return nil, err
	}

	ibdBlockMessage, ok := message.(*appmessage.MsgIBDBlock)
	if !ok {
		return nil, protocolerrors.Errorf(true, "received unexpected message type. "+
			"expected: %s, got: %s", appmessage.CmdIBDBlock, message.Command())
	}
	block := appmessage.MsgBlockToDomainBlock(ibdBlockMessage.MsgBlock)

	log.Debugf("Received pruning point block %s", consensushashing.BlockHash(block))

	return block, nil
}

func (flow *handleRelayInvsFlow) receiveAndInsertPruningPointUTXOSet(
	pruningPointHash *externalapi.DomainHash) (bool, error) {

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

			err := flow.Domain().Consensus().AppendImportedPruningPointUTXOs(domainOutpointAndUTXOEntryPairs)
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

			blockInsertionResult, err := flow.Domain().Consensus().ValidateAndInsertBlock(block)
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
