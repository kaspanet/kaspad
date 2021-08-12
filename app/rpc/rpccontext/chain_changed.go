package rpccontext

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// ConvertVirtualSelectedParentChainChangesToChainChangedNotificationMessage converts
// VirtualSelectedParentChainChanges to VirtualSelectedParentChainChangedNotificationMessage
func (ctx *Context) ConvertVirtualSelectedParentChainChangesToChainChangedNotificationMessage(
	selectedParentChainChanges *externalapi.SelectedChainPath) (*appmessage.VirtualSelectedParentChainChangedNotificationMessage, error) {

	removedChainBlockHashes := make([]string, len(selectedParentChainChanges.Removed))
	for i, removed := range selectedParentChainChanges.Removed {
		removedChainBlockHashes[i] = removed.String()
	}

	addedChainBlocks := make([]string, len(selectedParentChainChanges.Added))
	for i, added := range selectedParentChainChanges.Added {
		addedChainBlocks[i] = added.String()
	}

	return appmessage.NewVirtualSelectedParentChainChangedNotificationMessage(removedChainBlockHashes, addedChainBlocks), nil
}
