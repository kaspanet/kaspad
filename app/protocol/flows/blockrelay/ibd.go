package blockrelay

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/protocol/common"
	"github.com/kaspanet/kaspad/app/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/pkg/errors"
	"time"
)

func (flow *handleRelayInvsFlow) runIBDIfNotRunning(highHash *externalapi.DomainHash) error {
	wasIBDNotRunning := flow.TrySetIBDRunning()
	if !wasIBDNotRunning {
		log.Debugf("IBD is already running")
		return nil
	}
	defer flow.UnsetIBDRunning()

	log.Debugf("IBD started with peer %s and highHash %s", flow.peer, highHash)

	// Fetch all the headers if we don't already have them
	log.Debugf("Downloading headers up to %s", highHash)
	err := flow.syncHeaders(highHash)
	if err != nil {
		return err
	}
	log.Debugf("Finished downloading headers up to %s", highHash)

	// Fetch the UTXO set if we don't already have it
	log.Debugf("Downloading the UTXO set for %s", highHash)
	syncInfo, err := flow.Domain().Consensus().GetSyncInfo()
	if err != nil {
		return err
	}
	if syncInfo.State == externalapi.SyncStateAwaitingUTXOSet {
		found, err := flow.fetchMissingUTXOSet(syncInfo.IBDRootUTXOBlockHash)
		if err != nil {
			return err
		}
		if !found {
			log.Infof("Cannot download the UTXO set for %s "+
				"because the peer does not have its IBD root UTXO block", highHash)
			return nil
		}
	}
	log.Debugf("Finished downloading the UTXO set for %s", highHash)

	// Fetch the block bodies
	log.Debugf("Downloading block bodies up to %s", highHash)
	err = flow.syncMissingBlockBodies(highHash)
	if err != nil {
		return err
	}
	log.Debugf("Finished downloading block bodies up to %s", highHash)

	return nil
}

func (flow *handleRelayInvsFlow) syncHeaders(highHash *externalapi.DomainHash) error {
	log.Debugf("Trying to find highest shared chain block with peer %s with high hash %s", flow.peer, highHash)
	highestSharedBlockHash, err := flow.findHighestSharedBlockHash(highHash)
	if err != nil {
		return err
	}
	log.Debugf("Found highest shared chain block %s with peer %s", highestSharedBlockHash, flow.peer)

	return flow.downloadHeaders(highestSharedBlockHash, highHash)
}

func (flow *handleRelayInvsFlow) findHighestSharedBlockHash(highHash *externalapi.DomainHash) (
	lowHash *externalapi.DomainHash, err error) {

	lowHash = flow.Config().ActiveNetParams.GenesisHash
	currentHighHash := highHash

	for {
		err := flow.sendGetBlockLocator(lowHash, currentHighHash, 0)
		if err != nil {
			return nil, err
		}

		blockLocatorHashes, err := flow.receiveBlockLocator()
		if err != nil {
			return nil, err
		}

		// We check whether the locator's highest hash is in the local DAG.
		// If it is, return it. If it isn't, we need to narrow our
		// getBlockLocator request and try again.
		locatorHighHash := blockLocatorHashes[0]
		locatorHighHashInfo, err := flow.Domain().Consensus().GetBlockInfo(locatorHighHash)
		if err != nil {
			return nil, err
		}
		if locatorHighHashInfo.Exists {
			return locatorHighHash, nil
		}

		lowHash, currentHighHash, err = flow.Domain().Consensus().FindNextBlockLocatorBoundaries(blockLocatorHashes)
		if err != nil {
			return nil, err
		}
	}
}

func (flow *handleRelayInvsFlow) downloadHeaders(highestSharedBlockHash *externalapi.DomainHash,
	highHash *externalapi.DomainHash) error {

	err := flow.sendRequestHeaders(highestSharedBlockHash, highHash)
	if err != nil {
		return err
	}

	blocksReceived := 0
	for {
		msgBlockHeader, doneIBD, err := flow.receiveHeader()
		if err != nil {
			return err
		}
		if doneIBD {
			return nil
		}

		err = flow.processHeader(msgBlockHeader)
		if err != nil {
			return err
		}

		blocksReceived++
		if blocksReceived%ibdBatchSize == 0 {
			err = flow.outgoingRoute.Enqueue(appmessage.NewMsgRequestNextHeaders())
			if err != nil {
				return err
			}
		}
	}
}

func (flow *handleRelayInvsFlow) sendRequestHeaders(highestSharedBlockHash *externalapi.DomainHash,
	peerSelectedTipHash *externalapi.DomainHash) error {

	msgGetBlockInvs := appmessage.NewMsgRequstHeaders(highestSharedBlockHash, peerSelectedTipHash)
	return flow.outgoingRoute.Enqueue(msgGetBlockInvs)
}

func (flow *handleRelayInvsFlow) receiveHeader() (msgIBDBlock *appmessage.MsgBlockHeader, doneIBD bool, err error) {
	message, err := flow.dequeueIncomingMessageAndSkipInvs(common.DefaultTimeout)
	if err != nil {
		return nil, false, err
	}
	switch message := message.(type) {
	case *appmessage.MsgBlockHeader:
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
	err = flow.Domain().Consensus().ValidateAndInsertBlock(block)
	if err != nil {
		if !errors.As(err, &ruleerrors.RuleError{}) {
			return errors.Wrapf(err, "failed to process header %s during IBD", blockHash)
		}
		log.Infof("Rejected block header %s from %s during IBD: %s", blockHash, flow.peer, err)

		return protocolerrors.Wrapf(true, err, "got invalid block %s during IBD", blockHash)
	}
	return nil
}

func (flow *handleRelayInvsFlow) fetchMissingUTXOSet(ibdRootHash *externalapi.DomainHash) (bool, error) {
	err := flow.outgoingRoute.Enqueue(appmessage.NewMsgRequestIBDRootUTXOSetAndBlock(ibdRootHash))
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

	err = flow.Domain().Consensus().ValidateAndInsertBlock(block)
	if err != nil {
		blockHash := consensushashing.BlockHash(block)
		return false, protocolerrors.ConvertToBanningProtocolErrorIfRuleError(err, "got invalid block %s during IBD", blockHash)
	}

	err = flow.Domain().Consensus().SetPruningPointUTXOSet(utxoSet)
	if err != nil {
		return false, protocolerrors.ConvertToBanningProtocolErrorIfRuleError(err, "error with IBD root UTXO set")
	}

	return true, nil
}

func (flow *handleRelayInvsFlow) receiveIBDRootUTXOSetAndBlock() ([]byte, *externalapi.DomainBlock, bool, error) {
	message, err := flow.dequeueIncomingMessageAndSkipInvs(common.DefaultTimeout)
	if err != nil {
		return nil, nil, false, err
	}

	switch message := message.(type) {
	case *appmessage.MsgIBDRootUTXOSetAndBlock:
		return message.UTXOSet,
			appmessage.MsgBlockToDomainBlock(message.Block), true, nil
	case *appmessage.MsgIBDRootNotFound:
		return nil, nil, false, nil
	default:
		return nil, nil, false,
			protocolerrors.Errorf(true, "received unexpected message type. "+
				"expected: %s or %s, got: %s",
				appmessage.CmdIBDRootUTXOSetAndBlock, appmessage.CmdIBDRootNotFound, message.Command(),
			)
	}
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
			if *expectedHash != *blockHash {
				return protocolerrors.Errorf(true, "expected block %s but got %s", expectedHash, blockHash)
			}

			err = flow.Domain().Consensus().ValidateAndInsertBlock(block)
			if err != nil {
				return protocolerrors.ConvertToBanningProtocolErrorIfRuleError(err, "invalid block %s", blockHash)
			}
			err = flow.OnNewBlock(block)
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
