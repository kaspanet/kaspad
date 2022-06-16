package rpcclient

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	routerpkg "github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/pkg/errors"
)

// RegisterForVirtualSelectedParentChainChangedNotifications sends an RPC request respective to the function's name and returns the RPC server's response.
// Additionally, it starts listening for the appropriate notification using the given handler function
func (c *RPCClient) RegisterForVirtualSelectedParentChainChangedNotifications(includeAcceptedTransactionIDs bool,
	onChainChanged func(notification *appmessage.VirtualSelectedParentChainChangedNotificationMessage)) error {

	err := c.rpcRouter.outgoingRoute().Enqueue(
		appmessage.NewNotifyVirtualSelectedParentChainChangedRequestMessage(includeAcceptedTransactionIDs, rpccontext.DefaultNotificationId))
	if err != nil {
		return err
	}
	response, err := c.route(appmessage.CmdNotifyVirtualSelectedParentChainChangedResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return err
	}
	notifyChainChangedResponse := response.(*appmessage.NotifyVirtualSelectedParentChainChangedResponseMessage)
	if notifyChainChangedResponse.Error != nil {
		return c.convertRPCError(notifyChainChangedResponse.Error)
	}
	spawn("RegisterForVirtualSelectedParentChainChangedNotifications", func() {
		for {
			notification, err := c.route(appmessage.CmdVirtualSelectedParentChainChangedNotificationMessage).Dequeue()
			if err != nil {
				if errors.Is(err, routerpkg.ErrRouteClosed) {
					break
				}
				panic(err)
			}
			ChainChangedNotification := notification.(*appmessage.VirtualSelectedParentChainChangedNotificationMessage)
			onChainChanged(ChainChangedNotification)
		}
	})
	return nil
}
