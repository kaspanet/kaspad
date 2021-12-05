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
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/pkg/errors"
)

// orphanResolutionRange is the maximum amount of blockLocator hashes
// to search for known blocks. See isBlockInOrphanResolutionRange for
// further details
var orphanResolutionRange uint32 = 5

// RelayInvsContext is the interface for the context needed for the HandleRelayInvs flow.
type RelayInvsContext interface {
	Domain() domain.Domain
	Config() *config.Config
	OnNewBlock(block *externalapi.DomainBlock, blockInsertionResult *externalapi.BlockInsertionResult) error
	OnPruningPointUTXOSetOverride() error
	SharedRequestedBlocks() *SharedRequestedBlocks
	Broadcast(message appmessage.Message) error
	AddOrphan(orphanBlock *externalapi.DomainBlock)
	GetOrphanRoots(orphanHash *externalapi.DomainHash) ([]*externalapi.DomainHash, bool, error)
	IsOrphan(blockHash *externalapi.DomainHash) bool
	IsIBDRunning() bool
	TrySetIBDRunning(ibdPeer *peerpkg.Peer) bool
	UnsetIBDRunning()
	IsRecoverableError(err error) bool
}

type handleRelayInvsFlow struct {
	RelayInvsContext
	incomingRoute, outgoingRoute *router.Route
	peer                         *peerpkg.Peer
	invsQueue                    []*appmessage.MsgInvRelayBlock
}

// HandleRelayInvs listens to appmessage.MsgInvRelayBlock messages, requests their corresponding blocks if they
// are missing, adds them to the DAG and propagates them to the rest of the network.
func HandleRelayInvs(context RelayInvsContext, incomingRoute *router.Route, outgoingRoute *router.Route,
	peer *peerpkg.Peer) error {

	flow := &handleRelayInvsFlow{
		RelayInvsContext: context,
		incomingRoute:    incomingRoute,
		outgoingRoute:    outgoingRoute,
		peer:             peer,
		invsQueue:        make([]*appmessage.MsgInvRelayBlock, 0),
	}
	return flow.start()
}

func (flow *handleRelayInvsFlow) start() error {
	for {
		log.Debugf("Waiting for inv")
		inv, err := flow.readInv()
		if err != nil {
			return err
		}

		log.Debugf("Got relay inv for block %s", inv.Hash)

		blockInfo, err := flow.Domain().Consensus().GetBlockInfo(inv.Hash)
		if err != nil {
			return err
		}
		if blockInfo.Exists && blockInfo.BlockStatus != externalapi.StatusHeaderOnly {
			if blockInfo.BlockStatus == externalapi.StatusInvalid {
				return protocolerrors.Errorf(true, "sent inv of an invalid block %s",
					inv.Hash)
			}
			log.Debugf("Block %s already exists. continuing...", inv.Hash)
			continue
		}

		isGenesisVirtualSelectedParent, err := flow.isGenesisVirtualSelectedParent()
		if err != nil {
			return err
		}

		if flow.IsOrphan(inv.Hash) {
			if flow.Config().NetParams().DisallowDirectBlocksOnTopOfGenesis && !flow.Config().AllowSubmitBlockWhenNotSynced && isGenesisVirtualSelectedParent {
				log.Infof("Cannot process orphan %s for a node with only the genesis block. The node needs to IBD "+
					"to the recent pruning point before normal operation can resume.", inv.Hash)
				continue
			}

			log.Debugf("Block %s is a known orphan. Requesting its missing ancestors", inv.Hash)
			err := flow.AddOrphanRootsToQueue(inv.Hash)
			if err != nil {
				return err
			}
			continue
		}

		// Block relay is disabled during IBD
		if flow.IsIBDRunning() {
			log.Debugf("Got block %s while in IBD. continuing...", inv.Hash)
			continue
		}

		log.Debugf("Requesting block %s", inv.Hash)
		block, exists, err := flow.requestBlock(inv.Hash)
		if err != nil {
			return err
		}
		if exists {
			log.Debugf("Aborting requesting block %s because it already exists", inv.Hash)
			continue
		}

		err = flow.banIfBlockIsHeaderOnly(block)
		if err != nil {
			return err
		}

		if flow.Config().NetParams().DisallowDirectBlocksOnTopOfGenesis && !flow.Config().AllowSubmitBlockWhenNotSynced && !flow.Config().Devnet && flow.isChildOfGenesis(block) {
			log.Infof("Cannot process %s because it's a direct child of genesis.", consensushashing.BlockHash(block))
			continue
		}

		log.Debugf("Processing block %s", inv.Hash)
		missingParents, blockInsertionResult, err := flow.processBlock(block)
		if err != nil {
			if errors.Is(err, ruleerrors.ErrPrunedBlock) {
				log.Infof("Ignoring pruned block %s", inv.Hash)
				continue
			}

			if errors.Is(err, ruleerrors.ErrDuplicateBlock) {
				log.Infof("Ignoring duplicate block %s", inv.Hash)
				continue
			}
			return err
		}
		if len(missingParents) > 0 {
			log.Debugf("Block %s is orphan and has missing parents: %s", inv.Hash, missingParents)
			err := flow.processOrphan(block)
			if err != nil {
				return err
			}
			continue
		}

		log.Debugf("Relaying block %s", inv.Hash)
		err = flow.relayBlock(block)
		if err != nil {
			return err
		}
		log.Infof("Accepted block %s via relay", inv.Hash)
		err = flow.OnNewBlock(block, blockInsertionResult)
		if err != nil {
			return err
		}
	}
}

func (flow *handleRelayInvsFlow) banIfBlockIsHeaderOnly(block *externalapi.DomainBlock) error {
	if len(block.Transactions) == 0 {
		return protocolerrors.Errorf(true, "sent header of %s block where expected block with body",
			consensushashing.BlockHash(block))
	}

	return nil
}

func (flow *handleRelayInvsFlow) readInv() (*appmessage.MsgInvRelayBlock, error) {
	if len(flow.invsQueue) > 0 {
		var inv *appmessage.MsgInvRelayBlock
		inv, flow.invsQueue = flow.invsQueue[0], flow.invsQueue[1:]
		return inv, nil
	}

	msg, err := flow.incomingRoute.Dequeue()
	if err != nil {
		return nil, err
	}

	inv, ok := msg.(*appmessage.MsgInvRelayBlock)
	if !ok {
		return nil, protocolerrors.Errorf(true, "unexpected %s message in the block relay handleRelayInvsFlow while "+
			"expecting an inv message", msg.Command())
	}
	return inv, nil
}

func (flow *handleRelayInvsFlow) requestBlock(requestHash *externalapi.DomainHash) (*externalapi.DomainBlock, bool, error) {
	exists := flow.SharedRequestedBlocks().addIfNotExists(requestHash)
	if exists {
		return nil, true, nil
	}

	// In case the function returns earlier than expected, we want to make sure flow.SharedRequestedBlocks() is
	// clean from any pending blocks.
	defer flow.SharedRequestedBlocks().remove(requestHash)

	getRelayBlocksMsg := appmessage.NewMsgRequestRelayBlocks([]*externalapi.DomainHash{requestHash})
	err := flow.outgoingRoute.Enqueue(getRelayBlocksMsg)
	if err != nil {
		return nil, false, err
	}

	msgBlock, err := flow.readMsgBlock()
	if err != nil {
		return nil, false, err
	}

	block := appmessage.MsgBlockToDomainBlock(msgBlock)
	blockHash := consensushashing.BlockHash(block)
	if !blockHash.Equal(requestHash) {
		return nil, false, protocolerrors.Errorf(true, "got unrequested block %s", blockHash)
	}

	return block, false, nil
}

// readMsgBlock returns the next msgBlock in msgChan, and populates invsQueue with any inv messages that meanwhile arrive.
//
// Note: this function assumes msgChan can contain only appmessage.MsgInvRelayBlock and appmessage.MsgBlock messages.
func (flow *handleRelayInvsFlow) readMsgBlock() (msgBlock *appmessage.MsgBlock, err error) {
	for {
		message, err := flow.incomingRoute.DequeueWithTimeout(common.DefaultTimeout)
		if err != nil {
			return nil, err
		}

		switch message := message.(type) {
		case *appmessage.MsgInvRelayBlock:
			flow.invsQueue = append(flow.invsQueue, message)
		case *appmessage.MsgBlock:
			return message, nil
		default:
			return nil, errors.Errorf("unexpected message %s", message.Command())
		}
	}
}

func (flow *handleRelayInvsFlow) processBlock(block *externalapi.DomainBlock) ([]*externalapi.DomainHash, *externalapi.BlockInsertionResult, error) {
	blockHash := consensushashing.BlockHash(block)
	blockInsertionResult, err := flow.Domain().Consensus().ValidateAndInsertBlock(block, true)
	if err != nil {
		if !errors.As(err, &ruleerrors.RuleError{}) {
			return nil, nil, errors.Wrapf(err, "failed to process block %s", blockHash)
		}

		missingParentsError := &ruleerrors.ErrMissingParents{}
		if errors.As(err, missingParentsError) {
			return missingParentsError.MissingParentHashes, nil, nil
		}
		log.Warnf("Rejected block %s from %s: %s", blockHash, flow.peer, err)
		return nil, nil, protocolerrors.Wrapf(true, err, "got invalid block %s from relay", blockHash)
	}
	return nil, blockInsertionResult, nil
}

func (flow *handleRelayInvsFlow) relayBlock(block *externalapi.DomainBlock) error {
	blockHash := consensushashing.BlockHash(block)
	return flow.Broadcast(appmessage.NewMsgInvBlock(blockHash))
}

func (flow *handleRelayInvsFlow) processOrphan(block *externalapi.DomainBlock) error {
	blockHash := consensushashing.BlockHash(block)

	// Return if the block has been orphaned from elsewhere already
	if flow.IsOrphan(blockHash) {
		log.Debugf("Skipping orphan processing for block %s because it is already an orphan", blockHash)
		return nil
	}

	// Add the block to the orphan set if it's within orphan resolution range
	isBlockInOrphanResolutionRange, err := flow.isBlockInOrphanResolutionRange(blockHash)
	if err != nil {
		return err
	}
	if isBlockInOrphanResolutionRange {
		if flow.Config().NetParams().DisallowDirectBlocksOnTopOfGenesis && !flow.Config().AllowSubmitBlockWhenNotSynced {
			isGenesisVirtualSelectedParent, err := flow.isGenesisVirtualSelectedParent()
			if err != nil {
				return err
			}

			if isGenesisVirtualSelectedParent {
				log.Infof("Cannot process orphan %s for a node with only the genesis block. The node needs to IBD "+
					"to the recent pruning point before normal operation can resume.", blockHash)
				return nil
			}
		}

		log.Debugf("Block %s is within orphan resolution range. "+
			"Adding it to the orphan set", blockHash)
		flow.AddOrphan(block)
		log.Debugf("Requesting block %s missing ancestors", blockHash)
		return flow.AddOrphanRootsToQueue(blockHash)
	}

	// Start IBD unless we already are in IBD
	log.Debugf("Block %s is out of orphan resolution range. "+
		"Attempting to start IBD against it.", blockHash)
	return flow.runIBDIfNotRunning(block)
}

func (flow *handleRelayInvsFlow) isGenesisVirtualSelectedParent() (bool, error) {
	virtualSelectedParent, err := flow.Domain().Consensus().GetVirtualSelectedParent()
	if err != nil {
		return false, err
	}

	return virtualSelectedParent.Equal(flow.Config().NetParams().GenesisHash), nil
}

func (flow *handleRelayInvsFlow) isChildOfGenesis(block *externalapi.DomainBlock) bool {
	parents := block.Header.DirectParents()
	return len(parents) == 1 && parents[0].Equal(flow.Config().NetParams().GenesisHash)
}

// isBlockInOrphanResolutionRange finds out whether the given blockHash should be
// retrieved via the unorphaning mechanism or via IBD. This method sends a
// getBlockLocator request to the peer with a limit of orphanResolutionRange.
// In the response, if we know none of the hashes, we should retrieve the given
// blockHash via IBD. Otherwise, via unorphaning.
func (flow *handleRelayInvsFlow) isBlockInOrphanResolutionRange(blockHash *externalapi.DomainHash) (bool, error) {
	err := flow.sendGetBlockLocator(blockHash, orphanResolutionRange)
	if err != nil {
		return false, err
	}

	blockLocatorHashes, err := flow.receiveBlockLocator()
	if err != nil {
		return false, err
	}
	for _, blockLocatorHash := range blockLocatorHashes {
		blockInfo, err := flow.Domain().Consensus().GetBlockInfo(blockLocatorHash)
		if err != nil {
			return false, err
		}
		if blockInfo.Exists && blockInfo.BlockStatus != externalapi.StatusHeaderOnly {
			return true, nil
		}
	}
	return false, nil
}

func (flow *handleRelayInvsFlow) AddOrphanRootsToQueue(orphan *externalapi.DomainHash) error {
	orphanRoots, orphanExists, err := flow.GetOrphanRoots(orphan)
	if err != nil {
		return err
	}

	if !orphanExists {
		log.Infof("Orphan block %s was missing from the orphan pool while requesting for its roots. This "+
			"probably happened because it was randomly evicted immediately after it was added.", orphan)
	}

	log.Infof("Block %s has %d missing ancestors. Adding them to the invs queue...", orphan, len(orphanRoots))

	invMessages := make([]*appmessage.MsgInvRelayBlock, len(orphanRoots))
	for i, root := range orphanRoots {
		log.Debugf("Adding block %s missing ancestor %s to the invs queue", orphan, root)
		invMessages[i] = appmessage.NewMsgInvBlock(root)
	}

	flow.invsQueue = append(invMessages, flow.invsQueue...)
	return nil
}
