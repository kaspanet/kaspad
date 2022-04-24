package rpccontext

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// ConvertVirtualSelectedParentChainChangesToChainChangedNotificationMessage converts
// VirtualSelectedParentChainChanges to VirtualSelectedParentChainChangedNotificationMessage
func (ctx *Context) ConvertVirtualSelectedParentChainChangesToChainChangedNotificationMessage(
	selectedParentChainChanges *externalapi.SelectedChainPath, includeAcceptedTransactionIDs bool) (
	*appmessage.VirtualSelectedParentChainChangedNotificationMessage, error) {

	removedChainBlockHashes := make([]string, len(selectedParentChainChanges.Removed))
	for i, removed := range selectedParentChainChanges.Removed {
		removedChainBlockHashes[i] = removed.String()
	}

	addedChainBlocks := make([]string, len(selectedParentChainChanges.Added))
	for i, added := range selectedParentChainChanges.Added {
		addedChainBlocks[i] = added.String()
	}

	var acceptedTransactionIDs []*appmessage.AcceptedTransactionIDs
	if includeAcceptedTransactionIDs {
		var err error
		acceptedTransactionIDs, err = ctx.getAndConvertAcceptedTransactionIDs(selectedParentChainChanges)
		if err != nil {
			return nil, err
		}
	}

	return appmessage.NewVirtualSelectedParentChainChangedNotificationMessage(
		removedChainBlockHashes, addedChainBlocks, acceptedTransactionIDs), nil
}

func (ctx *Context) getAndConvertAcceptedTransactionIDs(selectedParentChainChanges *externalapi.SelectedChainPath) (
	[]*appmessage.AcceptedTransactionIDs, error) {

	acceptedTransactionIDs := make([]*appmessage.AcceptedTransactionIDs, len(selectedParentChainChanges.Added))

	for i, addedChainBlock := range selectedParentChainChanges.Added {
		blockAcceptanceData, err := ctx.Domain.Consensus().GetBlockAcceptanceData(addedChainBlock)
		if err != nil {
			return nil, err
		}
		acceptedTransactionIDs[i] = &appmessage.AcceptedTransactionIDs{
			AcceptingBlockHash:     addedChainBlock.String(),
			AcceptedTransactionIDs: nil,
		}
		for _, blockAcceptanceData := range blockAcceptanceData {
			for _, transactionAcceptanceData := range blockAcceptanceData.TransactionAcceptanceData {
				if transactionAcceptanceData.IsAccepted {
					acceptedTransactionIDs[i].AcceptedTransactionIDs =
						append(acceptedTransactionIDs[i].AcceptedTransactionIDs,
							transactionAcceptanceData.Transaction.ID.String())
				}
			}
		}
	}

	return acceptedTransactionIDs, nil
}
