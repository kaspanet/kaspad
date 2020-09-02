package rpcclient

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	routerpkg "github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/pkg/errors"
)

func (c *RPCClient) RegisterForBlockAddedNotifications(onBlockAdded func(header *appmessage.BlockHeader)) error {
	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewNotifyBlockAddedRequestMessage())
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
	spawn("RegisterForBlockAddedNotifications-blockAddedNotificationChan", func() {
		for {
			notification, err := c.route(appmessage.CmdBlockAddedNotificationMessage).Dequeue()
			if err != nil {
				if errors.Is(err, routerpkg.ErrRouteClosed) {
					break
				}
				panic(err)
			}
			blockAddedNotification := notification.(*appmessage.BlockAddedNotificationMessage)
			onBlockAdded(&blockAddedNotification.Block.Header)
		}
	})
	return nil
}
