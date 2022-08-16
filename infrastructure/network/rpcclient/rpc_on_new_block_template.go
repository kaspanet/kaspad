package rpcclient

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	routerpkg "github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/pkg/errors"
)

// RegisterForNewBlockTemplateNotifications sends an RPC request respective to the function's name and returns the RPC server's response.
// Additionally, it starts listening for the appropriate notification using the given handler function
func (c *RPCClient) RegisterForNewBlockTemplateNotifications(onNewBlockTemplate func(notification *appmessage.NewBlockTemplateNotificationMessage)) error {
	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewNotifyNewBlockTemplateRequestMessage(rpccontext.DefaultNotificationID))
	if err != nil {
		return err
	}
	response, err := c.route(appmessage.CmdNotifyNewBlockTemplateResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return err
	}
	notifyNewBlockTemplateResponse := response.(*appmessage.NotifyNewBlockTemplateResponseMessage)
	if notifyNewBlockTemplateResponse.Error != nil {
		return c.convertRPCError(notifyNewBlockTemplateResponse.Error)
	}
	spawn("RegisterForNewBlockTemplateNotifications", func() {
		for {
			notification, err := c.route(appmessage.CmdNewBlockTemplateNotificationMessage).Dequeue()
			if err != nil {
				if errors.Is(err, routerpkg.ErrRouteClosed) {
					break
				}
				panic(err)
			}
			NewBlockTemplateNotification := notification.(*appmessage.NewBlockTemplateNotificationMessage)
			onNewBlockTemplate(NewBlockTemplateNotification)
		}
	})
	return nil
}

// RegisterForNewBlockTemplateNotificationsWithID does the same as
// RegisterForNewBlockTemplateNotifications, but allows the client to specify an id
func (c *RPCClient) RegisterForNewBlockTemplateNotificationsWithID(onNewBlockTemplate func(notification *appmessage.NewBlockTemplateNotificationMessage), id string) error {
	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewNotifyNewBlockTemplateRequestMessage(id))
	if err != nil {
		return err
	}
	response, err := c.route(appmessage.CmdNotifyNewBlockTemplateResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return err
	}
	notifyNewBlockTemplateResponse := response.(*appmessage.NotifyNewBlockTemplateResponseMessage)
	if notifyNewBlockTemplateResponse.Error != nil {
		return c.convertRPCError(notifyNewBlockTemplateResponse.Error)
	}
	spawn("RegisterForNewBlockTemplateNotificationsWithID", func() {
		for {
			notification, err := c.route(appmessage.CmdNewBlockTemplateNotificationMessage).Dequeue()
			if err != nil {
				if errors.Is(err, routerpkg.ErrRouteClosed) {
					break
				}
				panic(err)
			}
			NewBlockTemplateNotification := notification.(*appmessage.NewBlockTemplateNotificationMessage)
			onNewBlockTemplate(NewBlockTemplateNotification)
		}
	})
	return nil
}
