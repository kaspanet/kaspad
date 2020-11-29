package blockrelay

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/protocol/common"
	"github.com/kaspanet/kaspad/app/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensusserialization"
	"github.com/pkg/errors"
)

func (flow *handleRelayInvsFlow) runIBD(highHash *externalapi.DomainHash) error {
	for {
		syncInfo, err := flow.Domain().Consensus().GetSyncInfo()
		if err != nil {
			return err
		}

		switch syncInfo.State {
		case externalapi.SyncStateHeadersFirst:
			err := flow.syncHeaders(highHash)
			if err != nil {
				return err
			}
		case externalapi.SyncStateMissingUTXOSet:
			found, err := flow.fetchMissingUTXOSet(syncInfo.IBDRootUTXOBlockHash)
			if err != nil {
				return err
			}

			if !found {
				return nil
			}
		case externalapi.SyncStateMissingBlockBodies:
			err := flow.syncMissingBlockBodies(highHash)
			if err != nil {
				return err
			}
		case externalapi.SyncStateRelay:
			return nil
		default:
			return errors.Errorf("unexpected state %s", syncInfo.State)
		}
	}
}

func (flow *handleRelayInvsFlow) syncHeaders(peerSelectedTipHash *externalapi.DomainHash) error {
	log.Debugf("Trying to find highest shared chain block with peer %s with selected tip %s", flow.peer, peerSelectedTipHash)
	highestSharedBlockHash, err := flow.findHighestSharedBlockHash(peerSelectedTipHash)
	if err != nil {
		return err
	}

	log.Debugf("Found highest shared chain block %s with peer %s", highestSharedBlockHash, flow.peer)

	return flow.downloadHeaders(highestSharedBlockHash, peerSelectedTipHash)
}

func (flow *handleRelayInvsFlow) syncMissingBlockBodies(peerSelectedTipHash *externalapi.DomainHash) error {
	hashes, err := flow.Domain().Consensus().GetMissingBlockBodyHashes(peerSelectedTipHash)
	if err != nil {
		return err
	}

	for offset := 0; offset < len(hashes); offset += appmessage.MaxRequestIBDBlocksHashes {
		var hashesToRequest []*externalapi.DomainHash
		if offset+appmessage.MaxRequestIBDBlocksHashes < len(hashes) {
			hashesToRequest = hashes[offset : offset+appmessage.MaxRequestIBDBlocksHashes]
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
			blockHash := consensusserialization.BlockHash(block)
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
		blockHash := consensusserialization.BlockHash(block)
		return false, protocolerrors.ConvertToBanningProtocolErrorIfRuleError(err, "got invalid block %s during IBD", blockHash)
	}

	err = flow.Domain().Consensus().SetPruningPointUTXOSet(utxoSet)
	if err != nil {
		return false, protocolerrors.ConvertToBanningProtocolErrorIfRuleError(err, "error with IBD root UTXO set")
	}

	return true, nil
}

func (flow *handleRelayInvsFlow) receiveIBDRootUTXOSetAndBlock() ([]byte, *externalapi.DomainBlock, bool, error) {
	message, err := flow.incomingRoute.DequeueWithTimeout(common.DefaultTimeout)
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

func (flow *handleRelayInvsFlow) findHighestSharedBlockHash(peerSelectedTipHash *externalapi.DomainHash) (
	lowHash *externalapi.DomainHash, err error) {

	lowHash = flow.Config().ActiveNetParams.GenesisHash
	highHash := peerSelectedTipHash

	for {
		err := flow.sendGetBlockLocator(lowHash, highHash)
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

		highHash, lowHash, err = flow.Domain().Consensus().FindNextBlockLocatorBoundaries(blockLocatorHashes)
		if err != nil {
			return nil, err
		}
	}
}

func (flow *handleRelayInvsFlow) sendGetBlockLocator(lowHash *externalapi.DomainHash, highHash *externalapi.DomainHash) error {
	msgGetBlockLocator := appmessage.NewMsgRequestBlockLocator(highHash, lowHash)
	return flow.outgoingRoute.Enqueue(msgGetBlockLocator)
}

func (flow *handleRelayInvsFlow) receiveBlockLocator() (blockLocatorHashes []*externalapi.DomainHash, err error) {
	message, err := flow.incomingRoute.DequeueWithTimeout(common.DefaultTimeout)
	if err != nil {
		return nil, err
	}
	msgBlockLocator, ok := message.(*appmessage.MsgBlockLocator)
	if !ok {
		return nil,
			protocolerrors.Errorf(true, "received unexpected message type. "+
				"expected: %s, got: %s", appmessage.CmdBlockLocator, message.Command())
	}
	return msgBlockLocator.BlockLocatorHashes, nil
}

func (flow *handleRelayInvsFlow) downloadHeaders(highestSharedBlockHash *externalapi.DomainHash,
	peerSelectedTipHash *externalapi.DomainHash) error {

	err := flow.sendRequestHeaders(highestSharedBlockHash, peerSelectedTipHash)
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
	message, err := flow.incomingRoute.DequeueWithTimeout(common.DefaultTimeout)
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

	blockHash := consensusserialization.BlockHash(block)
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
