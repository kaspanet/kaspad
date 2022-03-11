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

	/*
		Algorithm:
			Request full selected chain block locator from syncer
			Find the highest block which we know
			Repeat the locator step over the new range until finding max(past(syncee) \cap chain(syncer))
	*/

	// Empty hashes indicate that the full chain is queried
	locatorHashes, err := flow.getSyncerChainBlockLocator(nil, nil)
	if err != nil {
		return err
	}
	if len(locatorHashes) == 0 {
		return protocolerrors.Errorf(true, "Expecting initial syncer chain block locator "+
			"to contain at least one element")
	}
	syncerHeaderSelectedTipHash := locatorHashes[0]
	var highestKnownSyncerChainHash *externalapi.DomainHash
	for {
		var lowestUnknownSyncerChainHash, currentHighestKnownSyncerChainHash *externalapi.DomainHash
		for _, syncerChainHash := range locatorHashes {
			info, err := flow.Domain().Consensus().GetBlockInfo(syncerChainHash)
			if err != nil {
				return err
			}
			if info.Exists {
				currentHighestKnownSyncerChainHash = syncerChainHash
				break
			}
			lowestUnknownSyncerChainHash = syncerChainHash
		}
		// No shared block, break
		if currentHighestKnownSyncerChainHash == nil {
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
			currentHighestKnownSyncerChainHash)
		if err != nil {
			return err
		}
		if len(locatorHashes) == 2 {
			if !locatorHashes[0].Equal(lowestUnknownSyncerChainHash) ||
				!locatorHashes[1].Equal(currentHighestKnownSyncerChainHash) {
				return protocolerrors.Errorf(true, "Expecting the high and low "+
					"hashes to match the locatorHashes if len(locatorHashes) is 2")
			}
			// We found our search target
			highestKnownSyncerChainHash = currentHighestKnownSyncerChainHash
			break
		}
		if len(locatorHashes) == 0 {
			// An empty locator signals that the syncer chain was modified and no longer contains one of
			// the queried hashes, so we restart the search
			locatorHashes, err = flow.getSyncerChainBlockLocator(nil, nil)
			if err != nil {
				return err
			}
			if len(locatorHashes) == 0 {
				return protocolerrors.Errorf(true, "Expecting initial syncer chain block locator "+
					"to contain at least one element")
			}
			// Reset syncer's header selected tip
			syncerHeaderSelectedTipHash = locatorHashes[0]
		}
	}

	log.Debugf("Found highest known syncer chain block %s from peer %s",
		highestKnownSyncerChainHash, flow.peer)

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

		// TODO: need DAA score of syncerHeaderSelectedTipHash
		err = flow.syncPruningPointFutureHeaders(
			flow.Domain().Consensus(),
			syncerHeaderSelectedTipHash, highestKnownSyncerChainHash, relayBlockHash, block.Header.DAAScore())
		if err != nil {
			return err
		}
	}

	err = flow.syncMissingBlockBodies(relayBlockHash)
	if err != nil {
		return err
	}

	log.Debugf("Finished syncing blocks up to %s", relayBlockHash)
	isFinishedSuccessfully = true
	return nil
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
	log.Infof("IBD finished %s", successString)
}

func (flow *handleIBDFlow) getSyncerChainBlockLocator(
	highHash, lowHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {

	requestIbdChainBlockLocatorMessage := appmessage.NewMsgIBDRequestChainBlockLocator(highHash, lowHash)
	err := flow.outgoingRoute.Enqueue(requestIbdChainBlockLocatorMessage)
	if err != nil {
		return nil, err
	}
	message, err := flow.incomingRoute.DequeueWithTimeout(common.DefaultTimeout)
	if err != nil {
		return nil, err
	}
	switch message := message.(type) {
	case *appmessage.MsgIBDChainBlockLocator:
		return message.BlockLocatorHashes, nil
	default:
		return nil, protocolerrors.Errorf(true, "received unexpected message type. "+
			"expected: %s, got: %s", appmessage.CmdIBDChainBlockLocator, message.Command())
	}
}

func (flow *handleIBDFlow) syncPruningPointFutureHeaders(consensus externalapi.Consensus,
	syncerHeaderSelectedTipHash, highestKnownSyncerChainHash, relayBlockHash *externalapi.DomainHash,
	highBlockDAAScore uint64) error {

	log.Infof("Downloading headers from %s", flow.peer)

	err := flow.sendRequestHeaders(highestKnownSyncerChainHash, syncerHeaderSelectedTipHash)
	if err != nil {
		return err
	}

	highestSharedBlockHeader, err := consensus.GetBlockHeader(highestKnownSyncerChainHash)
	if err != nil {
		return err
	}
	progressReporter := newIBDProgressReporter(highestSharedBlockHeader.DAAScore(), highBlockDAAScore, "block headers")

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
				// Finished downloading syncer selected tip blocks,
				// check if we already have the triggering relayBlockHash
				relayBlockInfo, err := consensus.GetBlockInfo(relayBlockHash)
				if err != nil {
					return err
				}
				if !relayBlockInfo.Exists {
					// Send a special header request for the past diff. This is expected to be a small,
					// as it is bounded to the size of virtual's mergeset
					err = flow.sendRequestPastDiff(syncerHeaderSelectedTipHash, relayBlockHash)
					if err != nil {
						return err
					}
					pastDiffHeadersMessage, pastDiffDone, err := flow.receiveHeaders()
					if err != nil {
						return err
					}
					if !pastDiffDone {
						return protocolerrors.Errorf(true,
							"Expected only one past diff header chunk for past(%s) setminus past(%s)",
							syncerHeaderSelectedTipHash, relayBlockHash)
					}
					for _, header := range pastDiffHeadersMessage.BlockHeaders {
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
						"highHash block %s from peer %s during block download", relayBlockHash, flow.peer)
				}
				return nil
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

func (flow *handleIBDFlow) sendRequestPastDiff(
	syncerHeaderSelectedTipHash, relayBlockHash *externalapi.DomainHash) error {

	msgGetPastDiff := appmessage.NewMsgRequestPastDiff(syncerHeaderSelectedTipHash, relayBlockHash)
	return flow.outgoingRoute.Enqueue(msgGetPastDiff)
}

func (flow *handleIBDFlow) sendRequestHeaders(
	highestKnownSyncerChainHash, syncerHeaderSelectedTipHash *externalapi.DomainHash) error {

	msgGetBlockInvs := appmessage.NewMsgRequstHeaders(highestKnownSyncerChainHash, syncerHeaderSelectedTipHash)
	return flow.outgoingRoute.Enqueue(msgGetBlockInvs)
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
