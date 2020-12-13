package rpcclient

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	routerpkg "github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/pkg/errors"
)

// RegisterForUTXOsChangedNotifications sends an RPC request respective to the function's name and returns the RPC server's response.
// Additionally, it starts listening for the appropriate notification using the given handler function
func (c *RPCClient) RegisterForUTXOsChangedNotifications(addresses []string,
	onUTXOsChanged func(notification *appmessage.UTXOsChangedNotificationMessage)) error {

	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewNotifyUTXOsChangedRequestMessage(addresses))
	if err != nil {
		return err
	}
	response, err := c.route(appmessage.CmdNotifyUTXOsChangedResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return err
	}
	notifyUTXOsChangedResponse := response.(*appmessage.NotifyUTXOsChangedResponseMessage)
	if notifyUTXOsChangedResponse.Error != nil {
		return c.convertRPCError(notifyUTXOsChangedResponse.Error)
	}
	spawn("RegisterForUTXOsChangedNotifications", func() {
		for {
			notification, err := c.route(appmessage.CmdUTXOsChangedNotificationMessage).Dequeue()
			if err != nil {
				if errors.Is(err, routerpkg.ErrRouteClosed) {
					break
				}
				panic(err)
			}
			UTXOsChangedNotification := notification.(*appmessage.UTXOsChangedNotificationMessage)
			onUTXOsChanged(UTXOsChangedNotification)
		}
	})
	return nil
}
