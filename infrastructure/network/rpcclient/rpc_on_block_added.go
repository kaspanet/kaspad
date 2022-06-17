package rpcclient

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	routerpkg "github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/pkg/errors"
)

// RegisterForBlockAddedNotifications sends an RPC request respective to the function's name and returns the RPC server's response.
// Additionally, it starts listening for the appropriate notification using the given handler function
func (c *RPCClient) RegisterForBlockAddedNotifications(onBlockAdded func(notification *appmessage.BlockAddedNotificationMessage)) error {
	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewNotifyBlockAddedRequestMessage(rpccontext.DefaultNotificationID))
	if err != nil {
		return err
	}
	response, err := c.route(appmessage.CmdNotifyBlockAddedResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return err
	}
	notifyBlockAddedResponse := response.(*appmessage.NotifyBlockAddedResponseMessage)
	if notifyBlockAddedResponse.Error != nil {
		return c.convertRPCError(notifyBlockAddedResponse.Error)
	}
	spawn("RegisterForBlockAddedNotifications", func() {
		for {
			notification, err := c.route(appmessage.CmdBlockAddedNotificationMessage).Dequeue()
			if err != nil {
				if errors.Is(err, routerpkg.ErrRouteClosed) {
					break
				}
				panic(err)
			}
			blockAddedNotification := notification.(*appmessage.BlockAddedNotificationMessage)
			onBlockAdded(blockAddedNotification)
		}
	})
	return nil
}

// RegisterForBlockAddedNotificationsWithID does the same as
// RegisterForBlockAddedNotifications, but allows the client to specify an id
func (c *RPCClient) RegisterForBlockAddedNotificationsWithID(onBlockAdded func(notification *appmessage.BlockAddedNotificationMessage), id string) error {
	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewNotifyBlockAddedRequestMessage(id))
	if err != nil {
		return err
	}
	response, err := c.route(appmessage.CmdNotifyBlockAddedResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return err
	}
	notifyBlockAddedResponse := response.(*appmessage.NotifyBlockAddedResponseMessage)
	if notifyBlockAddedResponse.Error != nil {
		return c.convertRPCError(notifyBlockAddedResponse.Error)
	}
	spawn("RegisterForBlockAddedNotificationsWithID", func() {
		for {
			notification, err := c.route(appmessage.CmdBlockAddedNotificationMessage).Dequeue()
			if err != nil {
				if errors.Is(err, routerpkg.ErrRouteClosed) {
					break
				}
				panic(err)
			}
			blockAddedNotification := notification.(*appmessage.BlockAddedNotificationMessage)
			onBlockAdded(blockAddedNotification)
		}
	})
	return nil
}
