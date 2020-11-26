package ibd

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/protocol/common"
	peerpkg "github.com/kaspanet/kaspad/app/protocol/peer"
	"github.com/kaspanet/kaspad/app/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/domain"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensusserialization"
	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/pkg/errors"
)

// HandleIBDContext is the interface for the context needed for the HandleIBD flow.
type HandleIBDContext interface {
	Domain() domain.Domain
	Config() *config.Config
	OnNewBlock(block *externalapi.DomainBlock) error
	FinishIBD() error
}

type handleIBDFlow struct {
	HandleIBDContext
	incomingRoute, outgoingRoute *router.Route
	peer                         *peerpkg.Peer
}

// HandleIBD waits for IBD start and handles it when IBD is triggered for this peer
func HandleIBD(context HandleIBDContext, incomingRoute *router.Route, outgoingRoute *router.Route,
	peer *peerpkg.Peer) error {

	flow := &handleIBDFlow{
		HandleIBDContext: context,
		incomingRoute:    incomingRoute,
		outgoingRoute:    outgoingRoute,
		peer:             peer,
	}
	return flow.start()
}

func (flow *handleIBDFlow) start() error {
	for {
		err := flow.runIBD()
		if err != nil {
			return err
		}
	}
}

func (flow *handleIBDFlow) runIBD() error {
	flow.peer.WaitForIBDStart()
	err := flow.ibdLoop()
	if err != nil {
		finishIBDErr := flow.FinishIBD()
		if finishIBDErr != nil {
			return finishIBDErr
		}
		return err
	}
	return flow.FinishIBD()
}

func (flow *handleIBDFlow) ibdLoop() error {
	for {
		syncInfo, err := flow.Domain().Consensus().GetSyncInfo()
		if err != nil {
			return err
		}

		switch syncInfo.State {
		case externalapi.SyncStateHeadersFirst:
			err := flow.syncHeaders()
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
			err := flow.syncMissingBlockBodies()
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

func (flow *handleIBDFlow) syncHeaders() error {
	peerSelectedTipHash := flow.peer.SelectedTipHash()
	log.Debugf("Trying to find highest shared chain block with peer %s with selected tip %s", flow.peer, peerSelectedTipHash)
	highestSharedBlockHash, err := flow.findHighestSharedBlockHash(peerSelectedTipHash)
	if err != nil {
		return err
	}

	log.Debugf("Found highest shared chain block %s with peer %s", highestSharedBlockHash, flow.peer)

	return flow.downloadHeaders(highestSharedBlockHash, peerSelectedTipHash)
}

func (flow *handleIBDFlow) syncMissingBlockBodies() error {
	hashes, err := flow.Domain().Consensus().GetMissingBlockBodyHashes(flow.peer.SelectedTipHash())
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
			if !expectedHash.Equal(blockHash) {
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

func (flow *handleIBDFlow) fetchMissingUTXOSet(ibdRootHash *externalapi.DomainHash) (bool, error) {
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

func (flow *handleIBDFlow) receiveIBDRootUTXOSetAndBlock() ([]byte, *externalapi.DomainBlock, bool, error) {
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

func (flow *handleIBDFlow) findHighestSharedBlockHash(peerSelectedTipHash *externalapi.DomainHash) (lowHash *externalapi.DomainHash,
	err error) {

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

func (flow *handleIBDFlow) sendGetBlockLocator(lowHash *externalapi.DomainHash, highHash *externalapi.DomainHash) error {

	msgGetBlockLocator := appmessage.NewMsgRequestBlockLocator(highHash, lowHash)
	return flow.outgoingRoute.Enqueue(msgGetBlockLocator)
}

func (flow *handleIBDFlow) receiveBlockLocator() (blockLocatorHashes []*externalapi.DomainHash, err error) {
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

func (flow *handleIBDFlow) downloadHeaders(highestSharedBlockHash *externalapi.DomainHash,
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

func (flow *handleIBDFlow) sendRequestHeaders(highestSharedBlockHash *externalapi.DomainHash,
	peerSelectedTipHash *externalapi.DomainHash) error {

	msgGetBlockInvs := appmessage.NewMsgRequstHeaders(highestSharedBlockHash, peerSelectedTipHash)
	return flow.outgoingRoute.Enqueue(msgGetBlockInvs)
}

func (flow *handleIBDFlow) receiveHeader() (msgIBDBlock *appmessage.MsgBlockHeader, doneIBD bool, err error) {
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

func (flow *handleIBDFlow) processHeader(msgBlockHeader *appmessage.MsgBlockHeader) error {
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
