package blockrelay

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// SendVirtualSelectedParentInvContext is the interface for the context needed for the SendVirtualSelectedParentInv flow.
type SendVirtualSelectedParentInvContext interface {
	Domain() domain.Domain
}

// SendVirtualSelectedParentInv sends a peer the selected parent hash of the virtual
func SendVirtualSelectedParentInv(context SendVirtualSelectedParentInvContext, outgoingRoute *router.Route) error {
	virtualSelectedParent, err := context.Domain().Consensus().GetVirtualSelectedParent()
	if err != nil {
		return err
	}
	virtualSelectedParentInv := appmessage.NewMsgInvBlock(virtualSelectedParent)
	return outgoingRoute.Enqueue(virtualSelectedParentInv)
}
