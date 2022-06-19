package rpcclient

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	routerpkg "github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/pkg/errors"
)

// RegisterPruningPointUTXOSetNotifications sends an RPC request respective to the function's name and returns the RPC server's response.
// Additionally, it starts listening for the appropriate notification using the given handler function
func (c *RPCClient) RegisterPruningPointUTXOSetNotifications(onPruningPointUTXOSetNotifications func(notification *appmessage.PruningPointUTXOSetOverrideNotificationMessage)) error {

	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewNotifyPruningPointUTXOSetOverrideRequestMessage(rpccontext.DefaultNotificationID))
	if err != nil {
		return err
	}
	response, err := c.route(appmessage.CmdNotifyPruningPointUTXOSetOverrideResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return err
	}
	notifyPruningPointUTXOSetOverrideResponse := response.(*appmessage.NotifyPruningPointUTXOSetOverrideResponseMessage)
	if notifyPruningPointUTXOSetOverrideResponse.Error != nil {
		return c.convertRPCError(notifyPruningPointUTXOSetOverrideResponse.Error)
	}
	spawn("RegisterPruningPointUTXOSetNotifications", func() {
		for {
			notification, err := c.route(appmessage.CmdPruningPointUTXOSetOverrideNotificationMessage).Dequeue()
			if err != nil {
				if errors.Is(err, routerpkg.ErrRouteClosed) {
					break
				}
				panic(err)
			}
			newPruningPointUTXOSetOverrideNotification := notification.(*appmessage.PruningPointUTXOSetOverrideNotificationMessage) // Sanity check the type
			onPruningPointUTXOSetNotifications(newPruningPointUTXOSetOverrideNotification)
		}
	})
	return nil
}

// UnregisterPruningPointUTXOSetNotifications sends an RPC request respective to the function's name and returns the RPC server's response.
// Additionally, it stops listening for the appropriate notification using the given handler function
func (c *RPCClient) UnregisterPruningPointUTXOSetNotifications() error {

	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewStopNotifyingPruningPointUTXOSetOverrideRequestMessage(rpccontext.DefaultNotificationID))
	if err != nil {
		return err
	}
	response, err := c.route(appmessage.CmdStopNotifyingPruningPointUTXOSetOverrideResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return err
	}
	stopNotifyPruningPointUTXOSetOverrideResponse := response.(*appmessage.StopNotifyingPruningPointUTXOSetOverrideResponseMessage)
	if stopNotifyPruningPointUTXOSetOverrideResponse.Error != nil {
		return c.convertRPCError(stopNotifyPruningPointUTXOSetOverrideResponse.Error)
	}
	return nil
}

// RegisterPruningPointUTXOSetNotificationsWithID does the same as
// RegisterPruningPointUTXOSetNotifications, but allows the client to specify an id
func (c *RPCClient) RegisterPruningPointUTXOSetNotificationsWithID(onPruningPointUTXOSetNotifications func(notification *appmessage.PruningPointUTXOSetOverrideNotificationMessage), id string) error {

	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewNotifyPruningPointUTXOSetOverrideRequestMessage(id))
	if err != nil {
		return err
	}
	response, err := c.route(appmessage.CmdNotifyPruningPointUTXOSetOverrideResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return err
	}
	notifyPruningPointUTXOSetOverrideResponse := response.(*appmessage.NotifyPruningPointUTXOSetOverrideResponseMessage)
	if notifyPruningPointUTXOSetOverrideResponse.Error != nil {
		return c.convertRPCError(notifyPruningPointUTXOSetOverrideResponse.Error)
	}
	spawn("RegisterPruningPointUTXOSetNotificationsWithID", func() {
		for {
			notification, err := c.route(appmessage.CmdPruningPointUTXOSetOverrideNotificationMessage).Dequeue()
			if err != nil {
				if errors.Is(err, routerpkg.ErrRouteClosed) {
					break
				}
				panic(err)
			}
			newPruningPointUTXOSetOverrideNotification := notification.(*appmessage.PruningPointUTXOSetOverrideNotificationMessage) // Sanity check the type
			onPruningPointUTXOSetNotifications(newPruningPointUTXOSetOverrideNotification)
		}
	})
	return nil
}

// UnregisterPruningPointUTXOSetNotificationsWithID does the same as
// UnregisterPruningPointUTXOSetNotifications, but allows the client to specify an id
func (c *RPCClient) UnregisterPruningPointUTXOSetNotificationsWithID(id string) error {

	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewStopNotifyingPruningPointUTXOSetOverrideRequestMessage(id))
	if err != nil {
		return err
	}
	response, err := c.route(appmessage.CmdStopNotifyingPruningPointUTXOSetOverrideResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return err
	}
	stopNotifyPruningPointUTXOSetOverrideResponse := response.(*appmessage.StopNotifyingPruningPointUTXOSetOverrideResponseMessage)
	if stopNotifyPruningPointUTXOSetOverrideResponse.Error != nil {
		return c.convertRPCError(stopNotifyPruningPointUTXOSetOverrideResponse.Error)
	}
	return nil
}
