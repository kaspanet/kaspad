package rpcclient

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	routerpkg "github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/pkg/errors"
)

// RegisterForVirtualSelectedParentBlueScoreChangedNotifications sends an RPC request respective to the function's
// name and returns the RPC server's response. Additionally, it starts listening for the appropriate notification
// using the given handler function
func (c *RPCClient) RegisterForVirtualSelectedParentBlueScoreChangedNotifications(
	onVirtualSelectedParentBlueScoreChanged func(notification *appmessage.VirtualSelectedParentBlueScoreChangedNotificationMessage)) error {

	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewNotifyVirtualSelectedParentBlueScoreChangedRequestMessage(rpccontext.DefaultNotificationID))
	if err != nil {
		return err
	}
	response, err := c.route(appmessage.CmdNotifyVirtualSelectedParentBlueScoreChangedResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return err
	}
	notifyVirtualSelectedParentBlueScoreChangedResponse := response.(*appmessage.NotifyVirtualSelectedParentBlueScoreChangedResponseMessage)
	if notifyVirtualSelectedParentBlueScoreChangedResponse.Error != nil {
		return c.convertRPCError(notifyVirtualSelectedParentBlueScoreChangedResponse.Error)
	}
	spawn("RegisterForVirtualSelectedParentBlueScoreChangedNotifications", func() {
		for {
			notification, err := c.route(appmessage.CmdVirtualSelectedParentBlueScoreChangedNotificationMessage).Dequeue()
			if err != nil {
				if errors.Is(err, routerpkg.ErrRouteClosed) {
					break
				}
				panic(err)
			}
			VirtualSelectedParentBlueScoreChangedNotification := notification.(*appmessage.VirtualSelectedParentBlueScoreChangedNotificationMessage)
			onVirtualSelectedParentBlueScoreChanged(VirtualSelectedParentBlueScoreChangedNotification)
		}
	})
	return nil
}

// RegisterForVirtualSelectedParentBlueScoreChangedNotificationsWithID does the same as
// RegisterForVirtualSelectedParentBlueScoreChangedNotifications, but allows the client to specify an id
func (c *RPCClient) RegisterForVirtualSelectedParentBlueScoreChangedNotificationsWithID(
	onVirtualSelectedParentBlueScoreChanged func(notification *appmessage.VirtualSelectedParentBlueScoreChangedNotificationMessage), id string) error {

	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewNotifyVirtualSelectedParentBlueScoreChangedRequestMessage(id))
	if err != nil {
		return err
	}
	response, err := c.route(appmessage.CmdNotifyVirtualSelectedParentBlueScoreChangedResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return err
	}
	notifyVirtualSelectedParentBlueScoreChangedResponse := response.(*appmessage.NotifyVirtualSelectedParentBlueScoreChangedResponseMessage)
	if notifyVirtualSelectedParentBlueScoreChangedResponse.Error != nil {
		return c.convertRPCError(notifyVirtualSelectedParentBlueScoreChangedResponse.Error)
	}
	spawn("RegisterForVirtualSelectedParentBlueScoreChangedNotificationsWithID", func() {
		for {
			notification, err := c.route(appmessage.CmdVirtualSelectedParentBlueScoreChangedNotificationMessage).Dequeue()
			if err != nil {
				if errors.Is(err, routerpkg.ErrRouteClosed) {
					break
				}
				panic(err)
			}
			VirtualSelectedParentBlueScoreChangedNotification := notification.(*appmessage.VirtualSelectedParentBlueScoreChangedNotificationMessage)
			onVirtualSelectedParentBlueScoreChanged(VirtualSelectedParentBlueScoreChangedNotification)
		}
	})
	return nil
}
