package rpcclient

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	routerpkg "github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/pkg/errors"
)

// RegisterForChainChangedNotifications sends an RPC request respective to the function's name and returns the RPC server's response.
// Additionally, it starts listening for the appropriate notification using the given handler function
func (c *RPCClient) RegisterForChainChangedNotifications(onChainChanged func(notification *appmessage.ChainChangedNotificationMessage)) error {
	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewNotifyChainChangedRequestMessage())
	if err != nil {
		return err
	}
	response, err := c.route(appmessage.CmdNotifyChainChangedResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return err
	}
	notifyChainChangedResponse := response.(*appmessage.NotifyChainChangedResponseMessage)
	if notifyChainChangedResponse.Error != nil {
		return c.convertRPCError(notifyChainChangedResponse.Error)
	}
	spawn("RegisterForChainChangedNotifications", func() {
		for {
			notification, err := c.route(appmessage.CmdChainChangedNotificationMessage).Dequeue()
			if err != nil {
				if errors.Is(err, routerpkg.ErrRouteClosed) {
					break
				}
				panic(err)
			}
			ChainChangedNotification := notification.(*appmessage.ChainChangedNotificationMessage)
			onChainChanged(ChainChangedNotification)
		}
	})
	return nil
}
