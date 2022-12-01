package blockrelay

import (
	"fmt"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/protocol/common"
	"github.com/kaspanet/kaspad/app/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/pkg/errors"
	"time"
)

func (flow *handleIBDFlow) ibdWithHeadersProof(
	syncerHeaderSelectedTipHash, relayBlockHash *externalapi.DomainHash, highBlockDAAScore uint64) error {
	err := flow.Domain().InitStagingConsensusWithoutGenesis()
	if err != nil {
		return err
	}

	err = flow.downloadHeadersAndPruningUTXOSet(syncerHeaderSelectedTipHash, relayBlockHash, highBlockDAAScore)
	if err != nil {
		if !flow.IsRecoverableError(err) {
			return err
		}

		log.Infof("IBD with pruning proof from %s was unsuccessful. Deleting the staging consensus. (%s)", flow.peer, err)
		deleteStagingConsensusErr := flow.Domain().DeleteStagingConsensus()
		if deleteStagingConsensusErr != nil {
			return deleteStagingConsensusErr
		}

		return err
	}

	log.Infof("Header download stage of IBD with pruning proof completed successfully from %s. "+
		"Committing the staging consensus and deleting the previous obsolete one if such exists.", flow.peer)
	err = flow.Domain().CommitStagingConsensus()
	if err != nil {
		return err
	}

	err = flow.OnPruningPointUTXOSetOverride()
	if err != nil {
		return err
	}

	return nil
}

func (flow *handleIBDFlow) shouldSyncAndShouldDownloadHeadersProof(
	relayBlock *externalapi.DomainBlock,
	highestKnownSyncerChainHash *externalapi.DomainHash) (shouldDownload, shouldSync bool, err error) {

	var highestSharedBlockFound, isPruningPointInSharedBlockChain bool
	if highestKnownSyncerChainHash != nil {
		blockInfo, err := flow.Domain().Consensus().GetBlockInfo(highestKnownSyncerChainHash)
		if err != nil {
			return false, false, err
		}

		highestSharedBlockFound = blockInfo.HasBody()
		pruningPoint, err := flow.Domain().Consensus().PruningPoint()
		if err != nil {
			return false, false, err
		}

		isPruningPointInSharedBlockChain, err = flow.Domain().Consensus().IsInSelectedParentChainOf(
			pruningPoint, highestKnownSyncerChainHash)
		if err != nil {
			return false, false, err
		}
	}
	// Note: in the case where `highestSharedBlockFound == true && isPruningPointInSharedBlockChain == false`
	// we might have here info which is relevant to finality conflict decisions. This should be taken into
	// account when we improve this aspect.
	if !highestSharedBlockFound || !isPruningPointInSharedBlockChain {
		hasMoreBlueWorkThanSelectedTipAndPruningDepthMoreBlueScore, err := flow.checkIfHighHashHasMoreBlueWorkThanSelectedTipAndPruningDepthMoreBlueScore(relayBlock)
		if err != nil {
			return false, false, err
		}

		if hasMoreBlueWorkThanSelectedTipAndPruningDepthMoreBlueScore {
			return true, true, nil
		}

		if highestKnownSyncerChainHash == nil {
			log.Infof("Stopping IBD since IBD from this node will cause a finality conflict")
			return false, false, nil
		}

		return false, true, nil
	}

	return false, true, nil
}

func (flow *handleIBDFlow) checkIfHighHashHasMoreBlueWorkThanSelectedTipAndPruningDepthMoreBlueScore(relayBlock *externalapi.DomainBlock) (bool, error) {
	virtualSelectedParent, err := flow.Domain().Consensus().GetVirtualSelectedParent()
	if err != nil {
		return false, err
	}

	virtualSelectedTipInfo, err := flow.Domain().Consensus().GetBlockInfo(virtualSelectedParent)
	if err != nil {
		return false, err
	}

	if relayBlock.Header.BlueScore() < virtualSelectedTipInfo.BlueScore+flow.Config().NetParams().PruningDepth() {
		return false, nil
	}

	return relayBlock.Header.BlueWork().Cmp(virtualSelectedTipInfo.BlueWork) > 0, nil
}

func (flow *handleIBDFlow) syncAndValidatePruningPointProof() (*externalapi.DomainHash, error) {
	log.Infof("Downloading the pruning point proof from %s", flow.peer)
	err := flow.outgoingRoute.Enqueue(appmessage.NewMsgRequestPruningPointProof())
	if err != nil {
		return nil, err
	}
	message, err := flow.incomingRoute.DequeueWithTimeout(10 * time.Minute)
	if err != nil {
		return nil, err
	}
	pruningPointProofMessage, ok := message.(*appmessage.MsgPruningPointProof)
	if !ok {
		return nil, protocolerrors.Errorf(true, "received unexpected message type. "+
			"expected: %s, got: %s", appmessage.CmdPruningPointProof, message.Command())
	}
	pruningPointProof := appmessage.MsgPruningPointProofToDomainPruningPointProof(pruningPointProofMessage)
	err = flow.Domain().Consensus().ValidatePruningPointProof(pruningPointProof)
	if err != nil {
		if errors.As(err, &ruleerrors.RuleError{}) {
			return nil, protocolerrors.Wrapf(true, err, "pruning point proof validation failed")
		}
		return nil, err
	}

	err = flow.Domain().StagingConsensus().ApplyPruningPointProof(pruningPointProof)
	if err != nil {
		return nil, err
	}

	return consensushashing.HeaderHash(pruningPointProof.Headers[0][len(pruningPointProof.Headers[0])-1]), nil
}

func (flow *handleIBDFlow) downloadHeadersAndPruningUTXOSet(
	syncerHeaderSelectedTipHash, relayBlockHash *externalapi.DomainHash,
	highBlockDAAScore uint64) error {

	proofPruningPoint, err := flow.syncAndValidatePruningPointProof()
	if err != nil {
		return err
	}

	err = flow.syncPruningPointsAndPruningPointAnticone(proofPruningPoint)
	if err != nil {
		return err
	}

	// TODO: Remove this condition once there's more proper way to check finality violation
	// in the headers proof.
	if proofPruningPoint.Equal(flow.Config().NetParams().GenesisHash) {
		return protocolerrors.Errorf(true, "the genesis pruning point violates finality")
	}

	err = flow.syncPruningPointFutureHeaders(flow.Domain().StagingConsensus(),
		syncerHeaderSelectedTipHash, proofPruningPoint, relayBlockHash, highBlockDAAScore)
	if err != nil {
		return err
	}

	log.Infof("Headers downloaded from peer %s", flow.peer)

	relayBlockInfo, err := flow.Domain().StagingConsensus().GetBlockInfo(relayBlockHash)
	if err != nil {
		return err
	}

	if !relayBlockInfo.Exists {
		return protocolerrors.Errorf(true, "the triggering IBD block was not sent")
	}

	err = flow.validatePruningPointFutureHeaderTimestamps()
	if err != nil {
		return err
	}

	log.Debugf("Syncing the current pruning point UTXO set")
	syncedPruningPointUTXOSetSuccessfully, err := flow.syncPruningPointUTXOSet(flow.Domain().StagingConsensus(), proofPruningPoint)
	if err != nil {
		return err
	}
	if !syncedPruningPointUTXOSetSuccessfully {
		log.Debugf("Aborting IBD because the pruning point UTXO set failed to sync")
		return nil
	}
	log.Debugf("Finished syncing the current pruning point UTXO set")
	return nil
}

func (flow *handleIBDFlow) syncPruningPointsAndPruningPointAnticone(proofPruningPoint *externalapi.DomainHash) error {
	log.Infof("Downloading the past pruning points and the pruning point anticone from %s", flow.peer)
	err := flow.outgoingRoute.Enqueue(appmessage.NewMsgRequestPruningPointAndItsAnticone())
	if err != nil {
		return err
	}

	err = flow.validateAndInsertPruningPoints(proofPruningPoint)
	if err != nil {
		return err
	}

	message, err := flow.incomingRoute.DequeueWithTimeout(common.DefaultTimeout)
	if err != nil {
		return err
	}

	msgTrustedData, ok := message.(*appmessage.MsgTrustedData)
	if !ok {
		return protocolerrors.Errorf(true, "received unexpected message type. "+
			"expected: %s, got: %s", appmessage.CmdTrustedData, message.Command())
	}

	pruningPointWithMetaData, done, err := flow.receiveBlockWithTrustedData()
	if err != nil {
		return err
	}

	if done {
		return protocolerrors.Errorf(true, "got `done` message before receiving the pruning point")
	}

	if !pruningPointWithMetaData.Block.Header.BlockHash().Equal(proofPruningPoint) {
		return protocolerrors.Errorf(true, "first block with trusted data is not the pruning point")
	}

	err = flow.processBlockWithTrustedData(flow.Domain().StagingConsensus(), pruningPointWithMetaData, msgTrustedData)
	if err != nil {
		return err
	}

	i := 0
	for ; ; i++ {
		blockWithTrustedData, done, err := flow.receiveBlockWithTrustedData()
		if err != nil {
			return err
		}

		if done {
			break
		}

		err = flow.processBlockWithTrustedData(flow.Domain().StagingConsensus(), blockWithTrustedData, msgTrustedData)
		if err != nil {
			return err
		}

		// We're using i+2 because we want to check if the next block will belong to the next batch, but we already downloaded
		// the pruning point outside the loop so we use i+2 instead of i+1.
		if (i+2)%ibdBatchSize == 0 {
			log.Infof("Downloaded %d blocks from the pruning point anticone", i+1)
			err := flow.outgoingRoute.Enqueue(appmessage.NewMsgRequestNextPruningPointAndItsAnticoneBlocks())
			if err != nil {
				return err
			}
		}
	}

	log.Infof("Finished downloading pruning point and its anticone from %s. Total blocks downloaded: %d", flow.peer, i+1)
	return nil
}

func (flow *handleIBDFlow) processBlockWithTrustedData(
	consensus externalapi.Consensus, block *appmessage.MsgBlockWithTrustedDataV4, data *appmessage.MsgTrustedData) error {

	blockWithTrustedData := &externalapi.BlockWithTrustedData{
		Block:        appmessage.MsgBlockToDomainBlock(block.Block),
		DAAWindow:    make([]*externalapi.TrustedDataDataDAAHeader, 0, len(block.DAAWindowIndices)),
		GHOSTDAGData: make([]*externalapi.BlockGHOSTDAGDataHashPair, 0, len(block.GHOSTDAGDataIndices)),
	}

	for _, index := range block.DAAWindowIndices {
		blockWithTrustedData.DAAWindow = append(blockWithTrustedData.DAAWindow, appmessage.TrustedDataDataDAABlockV4ToTrustedDataDataDAAHeader(data.DAAWindow[index]))
	}

	for _, index := range block.GHOSTDAGDataIndices {
		blockWithTrustedData.GHOSTDAGData = append(blockWithTrustedData.GHOSTDAGData, appmessage.GHOSTDAGHashPairToDomainGHOSTDAGHashPair(data.GHOSTDAGData[index]))
	}

	err := consensus.ValidateAndInsertBlockWithTrustedData(blockWithTrustedData, false)
	if err != nil {
		if errors.As(err, &ruleerrors.RuleError{}) {
			return protocolerrors.Wrapf(true, err, "failed validating block with trusted data")
		}
		return err
	}
	return nil
}

func (flow *handleIBDFlow) receiveBlockWithTrustedData() (*appmessage.MsgBlockWithTrustedDataV4, bool, error) {
	message, err := flow.incomingRoute.DequeueWithTimeout(common.DefaultTimeout)
	if err != nil {
		return nil, false, err
	}

	switch downCastedMessage := message.(type) {
	case *appmessage.MsgBlockWithTrustedDataV4:
		return downCastedMessage, false, nil
	case *appmessage.MsgDoneBlocksWithTrustedData:
		return nil, true, nil
	default:
		return nil, false,
			protocolerrors.Errorf(true, "received unexpected message type. "+
				"expected: %s or %s, got: %s",
				(&appmessage.MsgBlockWithTrustedData{}).Command(),
				(&appmessage.MsgDoneBlocksWithTrustedData{}).Command(),
				downCastedMessage.Command())
	}
}

func (flow *handleIBDFlow) receivePruningPoints() (*appmessage.MsgPruningPoints, error) {
	message, err := flow.incomingRoute.DequeueWithTimeout(common.DefaultTimeout)
	if err != nil {
		return nil, err
	}

	msgPruningPoints, ok := message.(*appmessage.MsgPruningPoints)
	if !ok {
		return nil,
			protocolerrors.Errorf(true, "received unexpected message type. "+
				"expected: %s, got: %s", appmessage.CmdPruningPoints, message.Command())
	}

	return msgPruningPoints, nil
}

func (flow *handleIBDFlow) validateAndInsertPruningPoints(proofPruningPoint *externalapi.DomainHash) error {
	currentPruningPoint, err := flow.Domain().Consensus().PruningPoint()
	if err != nil {
		return err
	}

	if currentPruningPoint.Equal(proofPruningPoint) {
		return protocolerrors.Errorf(true, "the proposed pruning point is the same as the current pruning point")
	}

	pruningPoints, err := flow.receivePruningPoints()
	if err != nil {
		return err
	}

	headers := make([]externalapi.BlockHeader, len(pruningPoints.Headers))
	for i, header := range pruningPoints.Headers {
		headers[i] = appmessage.BlockHeaderToDomainBlockHeader(header)
	}

	arePruningPointsViolatingFinality, err := flow.Domain().Consensus().ArePruningPointsViolatingFinality(headers)
	if err != nil {
		return err
	}

	if arePruningPointsViolatingFinality {
		// TODO: Find a better way to deal with finality conflicts.
		return protocolerrors.Errorf(false, "pruning points are violating finality")
	}

	lastPruningPoint := consensushashing.HeaderHash(headers[len(headers)-1])
	if !lastPruningPoint.Equal(proofPruningPoint) {
		return protocolerrors.Errorf(true, "the proof pruning point is not equal to the last pruning "+
			"point in the list")
	}

	err = flow.Domain().StagingConsensus().ImportPruningPoints(headers)
	if err != nil {
		return err
	}

	return nil
}

func (flow *handleIBDFlow) syncPruningPointUTXOSet(consensus externalapi.Consensus,
	pruningPoint *externalapi.DomainHash) (bool, error) {

	log.Infof("Checking if the suggested pruning point %s is compatible to the node DAG", pruningPoint)
	isValid, err := flow.Domain().StagingConsensus().IsValidPruningPoint(pruningPoint)
	if err != nil {
		return false, err
	}

	if !isValid {
		return false, protocolerrors.Errorf(true, "invalid pruning point %s", pruningPoint)
	}

	log.Info("Fetching the pruning point UTXO set")
	isSuccessful, err := flow.fetchMissingUTXOSet(consensus, pruningPoint)
	if err != nil {
		log.Infof("An error occurred while fetching the pruning point UTXO set. Stopping IBD. (%s)", err)
		return false, err
	}

	if !isSuccessful {
		log.Infof("Couldn't successfully fetch the pruning point UTXO set. Stopping IBD.")
		return false, nil
	}

	log.Info("Fetched the new pruning point UTXO set")
	return true, nil
}

func (flow *handleIBDFlow) fetchMissingUTXOSet(consensus externalapi.Consensus, pruningPointHash *externalapi.DomainHash) (succeed bool, err error) {
	defer func() {
		err := flow.Domain().StagingConsensus().ClearImportedPruningPointData()
		if err != nil {
			panic(fmt.Sprintf("failed to clear imported pruning point data: %s", err))
		}
	}()

	err = flow.outgoingRoute.Enqueue(appmessage.NewMsgRequestPruningPointUTXOSet(pruningPointHash))
	if err != nil {
		return false, err
	}

	receivedAll, err := flow.receiveAndInsertPruningPointUTXOSet(consensus, pruningPointHash)
	if err != nil {
		return false, err
	}
	if !receivedAll {
		return false, nil
	}

	err = flow.Domain().StagingConsensus().ValidateAndInsertImportedPruningPoint(pruningPointHash)
	if err != nil {
		// TODO: Find a better way to deal with finality conflicts.
		if errors.Is(err, ruleerrors.ErrSuggestedPruningViolatesFinality) {
			return false, nil
		}
		return false, protocolerrors.ConvertToBanningProtocolErrorIfRuleError(err, "error with pruning point UTXO set")
	}

	return true, nil
}
