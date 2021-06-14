package blockrelay

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/protocol/common"
	"github.com/kaspanet/kaspad/app/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/pkg/errors"
)

// syncHeaders attempts to sync headers from the peer. This method may fail
// because the peer and us have conflicting pruning points. In that case we
// return (false, nil) so that we may stop IBD gracefully.
func (flow *handleRelayInvsFlow) syncHeaders(highHash *externalapi.DomainHash) (bool, error) {
	log.Debugf("Trying to find highest shared chain block with peer %s with high hash %s", flow.peer, highHash)
	highestSharedBlockHash, highestSharedBlockFound, err := flow.findHighestSharedBlockHash(highHash)
	if err != nil {
		return false, err
	}
	log.Debugf("Found highest shared chain block %s with peer %s", highestSharedBlockHash, flow.peer)

	shouldDownloadHeadersProof, shouldSync, err := flow.shouldDownloadHeadersProof(highHash, highestSharedBlockHash, highestSharedBlockFound)
	if err != nil {
		return false, err
	}

	if !shouldSync {
		return false, nil
	}

	if !shouldDownloadHeadersProof {
		return true, nil
	}

	err = flow.Domain().CommitStagingConsensus()
	if err != nil {
		return false, err
	}

	err = flow.downloadHeadersAndProof(highestSharedBlockHash, highHash)
	if !errors.As(err, &protocolerrors.ProtocolError{}) {
		err := flow.Domain().DeleteStagingConsensus()
		if err != nil {
			return false, err
		}
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

func (flow *handleRelayInvsFlow) shouldDownloadHeadersProof(highHash, highestSharedBlockHash *externalapi.DomainHash,
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

	blockInfo, err := flow.Domain().Consensus().GetBlockInfo(highestSharedBlockHash)
	if err != nil {
		return false, false, err
	}

	virtualInfo, err := flow.Domain().Consensus().GetVirtualInfo()
	if err != nil {
		return false, false, err
	}

	if virtualInfo.BlueScore-blockInfo.BlueScore > flow.Config().NetParams().PruningDepth() {
		hasMoreBlueWorkThanSelectedTip, err := flow.checkIfHighHashHasMoreBlueWorkThanSelectedTip(highHash)
		if err != nil {
			return false, false, err
		}

		return hasMoreBlueWorkThanSelectedTip, true, nil
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
				"expected: %s, got: %s", msgBlockBlueWork.Command(), message.Command())
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
	panic("unimplemented")
}

func (flow *handleRelayInvsFlow) downloadHeadersAndProof(highestSharedBlockHash, highHash *externalapi.DomainHash) error {
	err := flow.downloadHeadersProof()
	if err != nil {
		return err
	}

	err = flow.downloadHeaders(highestSharedBlockHash, highHash)
	if err != nil {
		return err
	}

	// If the highHash has not been received, the peer is misbehaving
	highHashBlockInfo, err := flow.Domain().Consensus().GetBlockInfo(highHash)
	if err != nil {
		return err
	}
	if !highHashBlockInfo.Exists {
		return protocolerrors.Errorf(true, "did not receive "+
			"highHash header %s from peer %s during header download", highHash, flow.peer)
	}
	log.Debugf("Headers downloaded from peer %s", flow.peer)
	return nil
}
