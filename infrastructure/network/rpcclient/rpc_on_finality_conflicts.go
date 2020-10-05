package rpcclient

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	routerpkg "github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/pkg/errors"
)

// RegisterForFinalityConflictsNotifications sends an RPC request respective to the function's name and returns the RPC server's response.
// Additionally, it starts listening for the appropriate notification using the given handler function
func (c *RPCClient) RegisterForFinalityConflictsNotifications(
	onFinalityConflict func(notification *appmessage.FinalityConflictNotificationMessage),
	onFinalityConflictResolved func(notification *appmessage.FinalityConflictResolvedNotificationMessage)) error {

	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewNotifyFinalityConflictsRequestMessage())
	if err != nil {
		return err
	}
	response, err := c.route(appmessage.CmdNotifyFinalityConflictsResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return err
	}
	notifyFinalityConflictsResponse := response.(*appmessage.NotifyFinalityConflictsResponseMessage)
	if notifyFinalityConflictsResponse.Error != nil {
		return c.convertRPCError(notifyFinalityConflictsResponse.Error)
	}
	spawn("RegisterForFinalityConflictsNotifications-finalityConflict", func() {
		for {
			notification, err := c.route(appmessage.CmdFinalityConflictNotificationMessage).Dequeue()
			if err != nil {
				if errors.Is(err, routerpkg.ErrRouteClosed) {
					break
				}
				panic(err)
			}
			finalityConflictNotification := notification.(*appmessage.FinalityConflictNotificationMessage)
			onFinalityConflict(finalityConflictNotification)
		}
	})
	spawn("RegisterForFinalityConflictsNotifications-finalityConflictResolved", func() {
		for {
			notification, err := c.route(appmessage.CmdFinalityConflictResolvedNotificationMessage).Dequeue()
			if err != nil {
				if errors.Is(err, routerpkg.ErrRouteClosed) {
					break
				}
				panic(err)
			}
			finalityConflictResolvedNotification := notification.(*appmessage.FinalityConflictResolvedNotificationMessage)
			onFinalityConflictResolved(finalityConflictResolvedNotification)
		}
	})
	return nil
}
