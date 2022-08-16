package rpcclient

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	routerpkg "github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/pkg/errors"
)

// RegisterForVirtualDaaScoreChangedNotifications sends an RPC request respective to the function's
// name and returns the RPC server's response. Additionally, it starts listening for the appropriate notification
// using the given handler function
func (c *RPCClient) RegisterForVirtualDaaScoreChangedNotifications(
	onVirtualDaaScoreChanged func(notification *appmessage.VirtualDaaScoreChangedNotificationMessage)) error {

	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewNotifyVirtualDaaScoreChangedRequestMessage(rpccontext.DefaultNotificationID))
	if err != nil {
		return err
	}
	response, err := c.route(appmessage.CmdNotifyVirtualDaaScoreChangedResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return err
	}
	notifyVirtualDaaScoreChangedResponse := response.(*appmessage.NotifyVirtualDaaScoreChangedResponseMessage)
	if notifyVirtualDaaScoreChangedResponse.Error != nil {
		return c.convertRPCError(notifyVirtualDaaScoreChangedResponse.Error)
	}
	spawn("RegisterForVirtualDaaScoreChangedNotifications", func() {
		for {
			notification, err := c.route(appmessage.CmdVirtualDaaScoreChangedNotificationMessage).Dequeue()
			if err != nil {
				if errors.Is(err, routerpkg.ErrRouteClosed) {
					break
				}
				panic(err)
			}
			VirtualDaaScoreChangedNotification := notification.(*appmessage.VirtualDaaScoreChangedNotificationMessage)
			onVirtualDaaScoreChanged(VirtualDaaScoreChangedNotification)
		}
	})
	return nil
}

// RegisterForVirtualDaaScoreChangedNotificationsWithID does the same as
// RegisterForVirtualDaaScoreChangedNotifications, but allows the client to specify an id
func (c *RPCClient) RegisterForVirtualDaaScoreChangedNotificationsWithID(
	onVirtualDaaScoreChanged func(notification *appmessage.VirtualDaaScoreChangedNotificationMessage), id string) error {

	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewNotifyVirtualDaaScoreChangedRequestMessage(id))
	if err != nil {
		return err
	}
	response, err := c.route(appmessage.CmdNotifyVirtualDaaScoreChangedResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return err
	}
	notifyVirtualDaaScoreChangedResponse := response.(*appmessage.NotifyVirtualDaaScoreChangedResponseMessage)
	if notifyVirtualDaaScoreChangedResponse.Error != nil {
		return c.convertRPCError(notifyVirtualDaaScoreChangedResponse.Error)
	}
	spawn("RegisterForVirtualDaaScoreChangedNotificationsWithID", func() {
		for {
			notification, err := c.route(appmessage.CmdVirtualDaaScoreChangedNotificationMessage).Dequeue()
			if err != nil {
				if errors.Is(err, routerpkg.ErrRouteClosed) {
					break
				}
				panic(err)
			}
			VirtualDaaScoreChangedNotification := notification.(*appmessage.VirtualDaaScoreChangedNotificationMessage)
			onVirtualDaaScoreChanged(VirtualDaaScoreChangedNotification)
		}
	})
	return nil
}
