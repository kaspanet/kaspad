package blockrelay

import (
	"fmt"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/protocol/common"
	"github.com/kaspanet/kaspad/app/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/pkg/errors"
)

func (flow *handleRelayInvsFlow) ibdWithHeadersProof(highHash *externalapi.DomainHash) error {
	err := flow.Domain().InitStagingConsensus()
	if err != nil {
		return err
	}

	committed := false
	defer func() {
		if committed || !flow.IsRecoverableError(err) {
			return
		}

		err := flow.Domain().DeleteStagingConsensus()
		if err != nil {
			panic(err)
		}
	}()

	err = flow.downloadBlocksAndPruningUTXOSet(flow.Domain().StagingConsensus(), highHash)
	if err != nil {
		return err
	}

	err = flow.Domain().CommitStagingConsensus()
	if err != nil {
		return err
	}

	committed = true
	return nil
}

func (flow *handleRelayInvsFlow) shouldSyncAndShouldDownloadHeadersProof(highHash *externalapi.DomainHash,
	highestSharedBlockFound bool) (shouldDownload, shouldSync bool, err error) {

	if !highestSharedBlockFound {
		hasMoreBlueWorkThanSelectedTip, err := flow.checkIfHighHashHasMoreBlueWorkThanSelectedTip(highHash)
		if err != nil {
			return false, false, err
		}

		if hasMoreBlueWorkThanSelectedTip {
			return true, true, nil
		}

		return false, false, nil
	}

	return false, true, nil
}

func (flow *handleRelayInvsFlow) checkIfHighHashHasMoreBlueWorkThanSelectedTip(highHash *externalapi.DomainHash) (bool, error) {
	err := flow.outgoingRoute.Enqueue(appmessage.NewRequestBlockBlueWork(highHash))
	if err != nil {
		return false, err
	}

	message, err := flow.dequeueIncomingMessageAndSkipInvs(common.DefaultTimeout)
	if err != nil {
		return false, err
	}

	msgBlockBlueWork, ok := message.(*appmessage.MsgBlockBlueWork)
	if !ok {
		return false,
			protocolerrors.Errorf(true, "received unexpected message type. "+
				"expected: %s, got: %s", appmessage.CmdBlockBlueWork, message.Command())
	}

	headersSelectedTip, err := flow.Domain().Consensus().GetHeadersSelectedTip()
	if err != nil {
		return false, err
	}

	headersSelectedTipInfo, err := flow.Domain().Consensus().GetBlockInfo(headersSelectedTip)
	if err != nil {
		return false, err
	}

	return msgBlockBlueWork.BlueWork.Cmp(headersSelectedTipInfo.BlueWork) > 0, nil
}

func (flow *handleRelayInvsFlow) downloadHeadersProof() error {
	// TODO: Implement headers proof mechanism
	return nil
}

func (flow *handleRelayInvsFlow) downloadBlocksAndPruningUTXOSet(consensus externalapi.Consensus, highHash *externalapi.DomainHash) error {
	err := flow.downloadHeadersProof()
	if err != nil {
		return err
	}

	pruningPoint, err := flow.syncPruningPointAndItsAnticone(consensus)
	if err != nil {
		return err
	}

	err = flow.syncPruningPointFuture(consensus, pruningPoint, highHash, false)
	if err != nil {
		return err
	}

	log.Debugf("Blocks downloaded from peer %s", flow.peer)

	log.Debugf("Syncing the current pruning point UTXO set")
	syncedPruningPointUTXOSetSuccessfully, err := flow.syncPruningPointUTXOSet(consensus, pruningPoint)
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

func (flow *handleRelayInvsFlow) syncPruningPointAndItsAnticone(consensus externalapi.Consensus) (*externalapi.DomainHash, error) {
	log.Infof("Downloading pruning point and its anticone from %s", flow.peer)
	err := flow.outgoingRoute.Enqueue(appmessage.NewMsgRequestPruningPointAndItsAnticone())
	if err != nil {
		return nil, err
	}

	pruningPoint, _, err := flow.receiveBlockWithMetaData(true)
	if err != nil {
		return nil, err
	}

	err = flow.processBlockWithMetaData(consensus, pruningPoint)
	if err != nil {
		return nil, err
	}

	for {
		blockWithMetaData, done, err := flow.receiveBlockWithMetaData(false)
		if err != nil {
			return nil, err
		}

		if done {
			break
		}

		err = flow.processBlockWithMetaData(consensus, blockWithMetaData)
		if err != nil {
			return nil, err
		}
	}

	log.Infof("Finished downloading pruning point and its anticone from %s", flow.peer)
	return pruningPoint.Block.Header.BlockHash(), nil
}

func (flow *handleRelayInvsFlow) processBlockWithMetaData(
	consensus externalapi.Consensus, block *appmessage.MsgBlockWithMetaData) error {

	_, err := consensus.ValidateAndInsertBlockWithMetaData(appmessage.BlockWithMetaDataToDomainBlockWithMetaData(block), false)
	return err
}

func (flow *handleRelayInvsFlow) receiveBlockWithMetaData(mustAccept bool) (*appmessage.MsgBlockWithMetaData, bool, error) {
	message, err := flow.dequeueIncomingMessageAndSkipInvs(common.DefaultTimeout)
	if err != nil {
		return nil, false, err
	}

	switch downCastedMessage := message.(type) {
	case *appmessage.MsgBlockWithMetaData:
		return downCastedMessage, false, nil
	case *appmessage.MsgDoneBlocksWithMetaData:
		if mustAccept {
			return nil, false,
				protocolerrors.Errorf(true, "received unexpected message type. "+
					"expected: %s, got: %s", downCastedMessage.Command(), message.Command())
		}
		return nil, true, nil
	default:
		return nil, false,
			protocolerrors.Errorf(true, "received unexpected message type. "+
				"expected: %s or %s, got: %s",
				(&appmessage.MsgBlockWithMetaData{}).Command(),
				(&appmessage.MsgDoneBlocksWithMetaData{}).Command(),
				downCastedMessage.Command())
	}
}

func (flow *handleRelayInvsFlow) syncPruningPointUTXOSet(consensus externalapi.Consensus, pruningPoint *externalapi.DomainHash) (bool, error) {
	log.Infof("Checking if the suggested pruning point %s is compatible to the node DAG", pruningPoint)
	isValid, err := flow.Domain().StagingConsensus().IsValidPruningPoint(pruningPoint)
	if err != nil {
		return false, err
	}

	if !isValid {
		log.Infof("The suggested pruning point %s is incompatible to this node DAG, so stopping IBD with this"+
			" peer", pruningPoint)
		return false, nil
	}

	log.Info("Fetching the pruning point UTXO set")
	succeed, err := flow.fetchMissingUTXOSet(consensus, pruningPoint)
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

func (flow *handleRelayInvsFlow) fetchMissingUTXOSet(consensus externalapi.Consensus, pruningPointHash *externalapi.DomainHash) (succeed bool, err error) {
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

	err = flow.OnPruningPointUTXOSetOverride()
	if err != nil {
		return false, err
	}

	return true, nil
}
