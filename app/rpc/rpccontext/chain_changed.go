package rpccontext

import (
	"encoding/hex"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
)

// ConvertVirtualSelectedParentChainChangesToChainChangedNotificationMessage converts
// VirtualSelectedParentChainChanges to VirtualSelectedParentChainChangedNotificationMessage
func (ctx *Context) ConvertVirtualSelectedParentChainChangesToChainChangedNotificationMessage(
	selectedParentChainChanges *externalapi.SelectedParentChainChanges) (*appmessage.VirtualSelectedParentChainChangedNotificationMessage, error) {

	removedChainBlockHashes := make([]string, len(selectedParentChainChanges.Removed))
	for i, removed := range selectedParentChainChanges.Removed {
		removedChainBlockHashes[i] = hex.EncodeToString(removed[:])
	}

	addedChainBlocks := make([]*appmessage.ChainBlock, len(selectedParentChainChanges.Added))
	for i, added := range selectedParentChainChanges.Added {
		acceptanceData, err := ctx.Domain.Consensus().GetBlockAcceptanceData(added)
		if err != nil {
			return nil, err
		}
		acceptedBlocks := make([]*appmessage.AcceptedBlock, len(acceptanceData))
		for j, acceptedBlock := range acceptanceData {
			acceptedTransactionIDs := make([]string, len(acceptedBlock.TransactionAcceptanceData))
			for k, transaction := range acceptedBlock.TransactionAcceptanceData {
				transactionID := consensushashing.TransactionID(transaction.Transaction)
				acceptedTransactionIDs[k] = hex.EncodeToString(transactionID[:])
			}
			acceptedBlocks[j] = &appmessage.AcceptedBlock{
				Hash:                   hex.EncodeToString(acceptedBlock.BlockHash[:]),
				AcceptedTransactionIDs: acceptedTransactionIDs,
			}
		}

		addedChainBlocks[i] = &appmessage.ChainBlock{
			Hash:           hex.EncodeToString(added[:]),
			AcceptedBlocks: acceptedBlocks,
		}
	}

	return appmessage.NewVirtualSelectedParentChainChangedNotificationMessage(removedChainBlockHashes, addedChainBlocks), nil
}
