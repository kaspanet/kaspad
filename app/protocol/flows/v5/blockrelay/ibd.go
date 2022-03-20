package blockrelay

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/protocol/common"
	peerpkg "github.com/kaspanet/kaspad/app/protocol/peer"
	"github.com/kaspanet/kaspad/app/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/domain"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/pkg/errors"
	"time"
)

// IBDContext is the interface for the context needed for the HandleIBD flow.
type IBDContext interface {
	Domain() domain.Domain
	Config() *config.Config
	OnNewBlock(block *externalapi.DomainBlock, virtualChangeSet *externalapi.VirtualChangeSet) error
	OnVirtualChange(virtualChangeSet *externalapi.VirtualChangeSet) error
	OnPruningPointUTXOSetOverride() error
	IsIBDRunning() bool
	TrySetIBDRunning(ibdPeer *peerpkg.Peer) bool
	UnsetIBDRunning()
	IsRecoverableError(err error) bool
}

type handleIBDFlow struct {
	IBDContext
	incomingRoute, outgoingRoute *router.Route
	peer                         *peerpkg.Peer
}

// HandleIBD handles IBD
func HandleIBD(context IBDContext, incomingRoute *router.Route, outgoingRoute *router.Route,
	peer *peerpkg.Peer) error {

	flow := &handleIBDFlow{
		IBDContext:    context,
		incomingRoute: incomingRoute,
		outgoingRoute: outgoingRoute,
		peer:          peer,
	}
	return flow.start()
}

func (flow *handleIBDFlow) start() error {
	for {
		// Wait for IBD requests triggered by other flows
		block, ok := <-flow.peer.IBDRequestChannel()
		if !ok {
			return nil
		}
		err := flow.runIBDIfNotRunning(block)
		if err != nil {
			return err
		}
	}
}

func (flow *handleIBDFlow) runIBDIfNotRunning(block *externalapi.DomainBlock) error {
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

	relayBlockHash := consensushashing.BlockHash(block)

	log.Debugf("IBD started with peer %s and relayBlockHash %s", flow.peer, relayBlockHash)
	log.Debugf("Syncing blocks up to %s", relayBlockHash)
	log.Debugf("Trying to find highest known syncer chain block from peer %s with relay hash %s", flow.peer, relayBlockHash)

	syncerHeaderSelectedTipHash, highestKnownSyncerChainHash, err := flow.negotiateMissingSyncerChainSegment()
	if err != nil {
		return err
	}

	shouldDownloadHeadersProof, shouldSync, err := flow.shouldSyncAndShouldDownloadHeadersProof(
		block, highestKnownSyncerChainHash)
	if err != nil {
		return err
	}

	if !shouldSync {
		return nil
	}

	if shouldDownloadHeadersProof {
		log.Infof("Starting IBD with headers proof")
		err := flow.ibdWithHeadersProof(syncerHeaderSelectedTipHash, relayBlockHash, block.Header.DAAScore())
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
					"to the recent pruning point before normal operation can resume.", relayBlockHash)
				return nil
			}
		}

		err = flow.syncPruningPointFutureHeaders(
			flow.Domain().Consensus(),
			syncerHeaderSelectedTipHash, highestKnownSyncerChainHash, relayBlockHash, block.Header.DAAScore())
		if err != nil {
			return err
		}
	}

	// We start by syncing missing bodies over the syncer selected chain
	err = flow.syncMissingBlockBodies(syncerHeaderSelectedTipHash)
	if err != nil {
		return err
	}
	relayBlockInfo, err := flow.Domain().Consensus().GetBlockInfo(relayBlockHash)
	if err != nil {
		return err
	}
	// Relay block might be in the anticone of syncer selected tip, thus
	// check his chain for missing bodies as well.
	// Note: this operation can be slightly optimized to avoid the full chain search since relay block
	// is in syncer virtual mergeset which has bounded size.
	if relayBlockInfo.BlockStatus == externalapi.StatusHeaderOnly {
		err = flow.syncMissingBlockBodies(relayBlockHash)
		if err != nil {
			return err
		}
	}

	log.Debugf("Finished syncing blocks up to %s", relayBlockHash)
	isFinishedSuccessfully = true
	return nil
}

func (flow *handleIBDFlow) negotiateMissingSyncerChainSegment() (*externalapi.DomainHash, *externalapi.DomainHash, error) {
	/*
		Algorithm:
			Request full selected chain block locator from syncer
			Find the highest block which we know
			Repeat the locator step over the new range until finding max(past(syncee) \cap chain(syncer))
	*/

	// Empty hashes indicate that the full chain is queried
	locatorHashes, err := flow.getSyncerChainBlockLocator(nil, nil, common.DefaultTimeout)
	if err != nil {
		return nil, nil, err
	}
	if len(locatorHashes) == 0 {
		return nil, nil, protocolerrors.Errorf(true, "Expecting initial syncer chain block locator "+
			"to contain at least one element")
	}
	log.Debugf("IBD chain negotiation with peer %s started and received %d hashes (%s, %s)", flow.peer,
		len(locatorHashes), locatorHashes[0], locatorHashes[len(locatorHashes)-1])
	syncerHeaderSelectedTipHash := locatorHashes[0]
	var highestKnownSyncerChainHash *externalapi.DomainHash
	chainNegotiationRestartCounter := 0
	chainNegotiationZoomCounts := 0
	initialLocatorLen := len(locatorHashes)
	for {
		var lowestUnknownSyncerChainHash, currentHighestKnownSyncerChainHash *externalapi.DomainHash
		for _, syncerChainHash := range locatorHashes {
			info, err := flow.Domain().Consensus().GetBlockInfo(syncerChainHash)
			if err != nil {
				return nil, nil, err
			}
			if info.Exists {
				currentHighestKnownSyncerChainHash = syncerChainHash
				break
			}
			lowestUnknownSyncerChainHash = syncerChainHash
		}
		// No unknown blocks, break. Note this can only happen in the first iteration
		if lowestUnknownSyncerChainHash == nil {
			highestKnownSyncerChainHash = currentHighestKnownSyncerChainHash
			break
		}
		// No shared block, break
		if currentHighestKnownSyncerChainHash == nil {
			highestKnownSyncerChainHash = nil
			break
		}
		// No point in zooming further
		if len(locatorHashes) == 1 {
			highestKnownSyncerChainHash = currentHighestKnownSyncerChainHash
			break
		}
		// Zoom in
		locatorHashes, err = flow.getSyncerChainBlockLocator(
			lowestUnknownSyncerChainHash,
			currentHighestKnownSyncerChainHash, time.Second*10)
		if err != nil {
			return nil, nil, err
		}
		if len(locatorHashes) > 0 {
			if !locatorHashes[0].Equal(lowestUnknownSyncerChainHash) ||
				!locatorHashes[len(locatorHashes)-1].Equal(currentHighestKnownSyncerChainHash) {
				return nil, nil, protocolerrors.Errorf(true, "Expecting the high and low "+
					"hashes to match the locator bounds")
			}

			chainNegotiationZoomCounts++
			log.Debugf("IBD chain negotiation with peer %s zoomed in (%d) and received %d hashes (%s, %s)", flow.peer,
				chainNegotiationZoomCounts, len(locatorHashes), locatorHashes[0], locatorHashes[len(locatorHashes)-1])

			if len(locatorHashes) == 2 {
				// We found our search target
				highestKnownSyncerChainHash = currentHighestKnownSyncerChainHash
				break
			}

			if chainNegotiationZoomCounts > initialLocatorLen*2 {
				// Since the zoom-in always queries two consecutive entries in the previous locator, it is
				// expected to decrease in size at least every two iterations
				return nil, nil, protocolerrors.Errorf(true,
					"IBD chain negotiation: Number of zoom-in steps %d exceeded the upper bound of 2*%d",
					chainNegotiationZoomCounts, initialLocatorLen)
			}

		} else { // Empty locator signals a restart due to chain changes
			chainNegotiationZoomCounts = 0
			chainNegotiationRestartCounter++
			if chainNegotiationRestartCounter > 32 {
				return nil, nil, protocolerrors.Errorf(false,
					"IBD chain negotiation with syncer %s exceeded restart limit %d", flow.peer, chainNegotiationRestartCounter)
			}
			log.Warnf("IBD chain negotiation with syncer %s restarted %d times", flow.peer, chainNegotiationRestartCounter)

			// An empty locator signals that the syncer chain was modified and no longer contains one of
			// the queried hashes, so we restart the search. We use a shorter timeout here to avoid a timeout attack
			locatorHashes, err = flow.getSyncerChainBlockLocator(nil, nil, time.Second*10)
			if err != nil {
				return nil, nil, err
			}
			if len(locatorHashes) == 0 {
				return nil, nil, protocolerrors.Errorf(true, "Expecting initial syncer chain block locator "+
					"to contain at least one element")
			}
			log.Infof("IBD chain negotiation with peer %s restarted (%d) and received %d hashes (%s, %s)", flow.peer,
				chainNegotiationRestartCounter, len(locatorHashes), locatorHashes[0], locatorHashes[len(locatorHashes)-1])

			initialLocatorLen = len(locatorHashes)
			// Reset syncer's header selected tip
			syncerHeaderSelectedTipHash = locatorHashes[0]
		}
	}

	log.Debugf("Found highest known syncer chain block %s from peer %s",
		highestKnownSyncerChainHash, flow.peer)

	return syncerHeaderSelectedTipHash, highestKnownSyncerChainHash, nil
}

func (flow *handleIBDFlow) isGenesisVirtualSelectedParent() (bool, error) {
	virtualSelectedParent, err := flow.Domain().Consensus().GetVirtualSelectedParent()
	if err != nil {
		return false, err
	}

	return virtualSelectedParent.Equal(flow.Config().NetParams().GenesisHash), nil
}

func (flow *handleIBDFlow) logIBDFinished(isFinishedSuccessfully bool) {
	successString := "successfully"
	if !isFinishedSuccessfully {
		successString = "(interrupted)"
	}
	log.Infof("IBD with peer %s finished %s", flow.peer, successString)
}

func (flow *handleIBDFlow) getSyncerChainBlockLocator(
	highHash, lowHash *externalapi.DomainHash, timeout time.Duration) ([]*externalapi.DomainHash, error) {

	requestIbdChainBlockLocatorMessage := appmessage.NewMsgIBDRequestChainBlockLocator(highHash, lowHash)
	err := flow.outgoingRoute.Enqueue(requestIbdChainBlockLocatorMessage)
	if err != nil {
		return nil, err
	}
	message, err := flow.incomingRoute.DequeueWithTimeout(timeout)
	if err != nil {
		return nil, err
	}
	switch message := message.(type) {
	case *appmessage.MsgIBDChainBlockLocator:
		if len(message.BlockLocatorHashes) > 64 {
			return nil, protocolerrors.Errorf(true,
				"Got block locator of size %d>64 while expecting locator to have size "+
					"which is logarithmic in DAG size (which should never exceed 2^64)",
				len(message.BlockLocatorHashes))
		}
		return message.BlockLocatorHashes, nil
	default:
		return nil, protocolerrors.Errorf(true, "received unexpected message type. "+
			"expected: %s, got: %s", appmessage.CmdIBDChainBlockLocator, message.Command())
	}
}

func (flow *handleIBDFlow) syncPruningPointFutureHeaders(consensus externalapi.Consensus,
	syncerHeaderSelectedTipHash, highestKnownSyncerChainHash, relayBlockHash *externalapi.DomainHash,
	highBlockDAAScoreHint uint64) error {

	log.Infof("Downloading headers from %s", flow.peer)

	if highestKnownSyncerChainHash.Equal(syncerHeaderSelectedTipHash) {
		// No need to get syncer selected tip headers, so sync relay past and return
		return flow.syncMissingRelayPast(consensus, syncerHeaderSelectedTipHash, relayBlockHash)
	}

	err := flow.sendRequestHeaders(highestKnownSyncerChainHash, syncerHeaderSelectedTipHash)
	if err != nil {
		return err
	}

	highestSharedBlockHeader, err := consensus.GetBlockHeader(highestKnownSyncerChainHash)
	if err != nil {
		return err
	}
	progressReporter := newIBDProgressReporter(highestSharedBlockHeader.DAAScore(), highBlockDAAScoreHint, "block headers")

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
			if len(blockHeadersMessage.BlockHeaders) == 0 {
				// The syncer should have sent a done message if the search completed, and not an empty list
				errChan <- protocolerrors.Errorf(true, "Received an empty headers message from peer %s", flow.peer)
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
				return flow.syncMissingRelayPast(consensus, syncerHeaderSelectedTipHash, relayBlockHash)
			}
			for _, header := range ibdBlocksMessage.BlockHeaders {
				err = flow.processHeader(consensus, header)
				if err != nil {
					return err
				}
			}

			lastReceivedHeader := ibdBlocksMessage.BlockHeaders[len(ibdBlocksMessage.BlockHeaders)-1]
			progressReporter.reportProgress(len(ibdBlocksMessage.BlockHeaders), lastReceivedHeader.DAAScore)
		case err := <-errChan:
			return err
		}
	}
}

func (flow *handleIBDFlow) syncMissingRelayPast(consensus externalapi.Consensus, syncerHeaderSelectedTipHash *externalapi.DomainHash, relayBlockHash *externalapi.DomainHash) error {
	// Finished downloading syncer selected tip blocks,
	// check if we already have the triggering relayBlockHash
	relayBlockInfo, err := consensus.GetBlockInfo(relayBlockHash)
	if err != nil {
		return err
	}
	if !relayBlockInfo.Exists {
		// Send a special header request for the selected tip anticone. This is expected to
		// be a small set, as it is bounded to the size of virtual's mergeset.
		err = flow.sendRequestAnticone(syncerHeaderSelectedTipHash, relayBlockHash)
		if err != nil {
			return err
		}
		anticoneHeadersMessage, anticoneDone, err := flow.receiveHeaders()
		if err != nil {
			return err
		}
		if anticoneDone {
			return protocolerrors.Errorf(true,
				"Expected one anticone header chunk for past(%s) cap anticone(%s) but got zero",
				relayBlockHash, syncerHeaderSelectedTipHash)
		}
		_, anticoneDone, err = flow.receiveHeaders()
		if err != nil {
			return err
		}
		if !anticoneDone {
			return protocolerrors.Errorf(true,
				"Expected only one anticone header chunk for past(%s) cap anticone(%s)",
				relayBlockHash, syncerHeaderSelectedTipHash)
		}
		for _, header := range anticoneHeadersMessage.BlockHeaders {
			err = flow.processHeader(consensus, header)
			if err != nil {
				return err
			}
		}
	}

	// If the relayBlockHash has still not been received, the peer is misbehaving
	relayBlockInfo, err = consensus.GetBlockInfo(relayBlockHash)
	if err != nil {
		return err
	}
	if !relayBlockInfo.Exists {
		return protocolerrors.Errorf(true, "did not receive "+
			"relayBlockHash block %s from peer %s during block download", relayBlockHash, flow.peer)
	}
	return nil
}

func (flow *handleIBDFlow) sendRequestAnticone(
	syncerHeaderSelectedTipHash, relayBlockHash *externalapi.DomainHash) error {

	msgRequestAnticone := appmessage.NewMsgRequestAnticone(syncerHeaderSelectedTipHash, relayBlockHash)
	return flow.outgoingRoute.Enqueue(msgRequestAnticone)
}

func (flow *handleIBDFlow) sendRequestHeaders(
	highestKnownSyncerChainHash, syncerHeaderSelectedTipHash *externalapi.DomainHash) error {

	msgRequestHeaders := appmessage.NewMsgRequstHeaders(highestKnownSyncerChainHash, syncerHeaderSelectedTipHash)
	return flow.outgoingRoute.Enqueue(msgRequestHeaders)
}

func (flow *handleIBDFlow) receiveHeaders() (msgIBDBlock *appmessage.BlockHeadersMessage, doneHeaders bool, err error) {
	message, err := flow.incomingRoute.DequeueWithTimeout(common.DefaultTimeout)
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

func (flow *handleIBDFlow) processHeader(consensus externalapi.Consensus, msgBlockHeader *appmessage.MsgBlockHeader) error {
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

func (flow *handleIBDFlow) validatePruningPointFutureHeaderTimestamps() error {
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

func (flow *handleIBDFlow) receiveAndInsertPruningPointUTXOSet(
	consensus externalapi.Consensus, pruningPointHash *externalapi.DomainHash) (bool, error) {

	onEnd := logger.LogAndMeasureExecutionTime(log, "receiveAndInsertPruningPointUTXOSet")
	defer onEnd()

	receivedChunkCount := 0
	receivedUTXOCount := 0
	for {
		message, err := flow.incomingRoute.DequeueWithTimeout(common.DefaultTimeout)
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

func (flow *handleIBDFlow) syncMissingBlockBodies(highHash *externalapi.DomainHash) error {
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

	lowBlockHeader, err := flow.Domain().Consensus().GetBlockHeader(hashes[0])
	if err != nil {
		return err
	}
	highBlockHeader, err := flow.Domain().Consensus().GetBlockHeader(hashes[len(hashes)-1])
	if err != nil {
		return err
	}
	progressReporter := newIBDProgressReporter(lowBlockHeader.DAAScore(), highBlockHeader.DAAScore(), "blocks")
	highestProcessedDAAScore := lowBlockHeader.DAAScore()

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
			message, err := flow.incomingRoute.DequeueWithTimeout(common.DefaultTimeout)
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

			virtualChangeSet, err := flow.Domain().Consensus().ValidateAndInsertBlock(block, false)
			if err != nil {
				if errors.Is(err, ruleerrors.ErrDuplicateBlock) {
					log.Debugf("Skipping IBD Block %s as it has already been added to the DAG", blockHash)
					continue
				}
				return protocolerrors.ConvertToBanningProtocolErrorIfRuleError(err, "invalid block %s", blockHash)
			}
			err = flow.OnNewBlock(block, virtualChangeSet)
			if err != nil {
				return err
			}

			highestProcessedDAAScore = block.Header.DAAScore()
		}

		progressReporter.reportProgress(len(hashesToRequest), highestProcessedDAAScore)
	}

	return flow.resolveVirtual(highestProcessedDAAScore)
}

func (flow *handleIBDFlow) banIfBlockIsHeaderOnly(block *externalapi.DomainBlock) error {
	if len(block.Transactions) == 0 {
		return protocolerrors.Errorf(true, "sent header of %s block where expected block with body",
			consensushashing.BlockHash(block))
	}

	return nil
}

func (flow *handleIBDFlow) resolveVirtual(estimatedVirtualDAAScoreTarget uint64) error {
	virtualDAAScoreStart, err := flow.Domain().Consensus().GetVirtualDAAScore()
	if err != nil {
		return err
	}

	for i := 0; ; i++ {
		if i%10 == 0 {
			virtualDAAScore, err := flow.Domain().Consensus().GetVirtualDAAScore()
			if err != nil {
				return err
			}
			var percents int
			if estimatedVirtualDAAScoreTarget-virtualDAAScoreStart <= 0 {
				percents = 100
			} else {
				percents = int(float64(virtualDAAScore-virtualDAAScoreStart) / float64(estimatedVirtualDAAScoreTarget-virtualDAAScoreStart) * 100)
			}
			log.Infof("Resolving virtual. Estimated progress: %d%%", percents)
		}
		virtualChangeSet, isCompletelyResolved, err := flow.Domain().Consensus().ResolveVirtual()
		if err != nil {
			return err
		}

		err = flow.OnVirtualChange(virtualChangeSet)
		if err != nil {
			return err
		}

		if isCompletelyResolved {
			log.Infof("Resolved virtual")
			return nil
		}
	}
}
