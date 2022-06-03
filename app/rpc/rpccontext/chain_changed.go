package rpccontext

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
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

	const chunk = 1000
	position := 0

	for position < len(selectedParentChainChanges.Added) {
		var chainBlocksChunk []*externalapi.DomainHash
		if position+chunk > len(selectedParentChainChanges.Added) {
			chainBlocksChunk = selectedParentChainChanges.Added[position:]
		} else {
			chainBlocksChunk = selectedParentChainChanges.Added[position : position+chunk]
		}
		// We use chunks in order to avoid blocking consensus for too long
		chainBlocksAcceptanceData, err := ctx.Domain.Consensus().GetBlocksAcceptanceData(chainBlocksChunk)
		if err != nil {
			return nil, err
		}

		for i, addedChainBlock := range chainBlocksChunk {
			chainBlockAcceptanceData := chainBlocksAcceptanceData[i]
			acceptedTransactionIDs[position+i] = &appmessage.AcceptedTransactionIDs{
				AcceptingBlockHash:     addedChainBlock.String(),
				AcceptedTransactionIDs: nil,
			}
			for _, blockAcceptanceData := range chainBlockAcceptanceData {
				for _, transactionAcceptanceData := range blockAcceptanceData.TransactionAcceptanceData {
					if transactionAcceptanceData.IsAccepted {
						acceptedTransactionIDs[position+i].AcceptedTransactionIDs =
							append(acceptedTransactionIDs[position+i].AcceptedTransactionIDs,
								consensushashing.TransactionID(transactionAcceptanceData.Transaction).String())
					}
				}
			}
		}
		position += chunk
	}

	return acceptedTransactionIDs, nil
}
