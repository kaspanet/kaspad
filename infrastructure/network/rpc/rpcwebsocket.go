package rpc

//func (m *wsNotificationManager) NotifyFinalityConflict(violatingBlockHash *daghash.Hash) {
//	n := notificationFinalityConflict{
//		violatingBlockHash: violatingBlockHash,
//	}
//	// As NotifyFinalityConflict will be called by the DAG manager
//	// and the RPC server may no longer be running, use a select
//	// statement to unblock enqueuing the notification once the RPC
//	// server has begun shutting down.
//	select {
//	case m.queueNotification <- n:
//	case <-m.quit:
//	}
//}
//
//func (m *wsNotificationManager) NotifyFinalityConflictResolved(finalityBlockHash *daghash.Hash) {
//
//	n := notificationFinalityConflictResolved{
//		finalityBlockHash: finalityBlockHash,
//	}
//	// As NotifyFinalityConflictResolved will be called by the DAG manager
//	// and the RPC server may no longer be running, use a select
//	// statement to unblock enqueuing the notification once the RPC
//	// server has begun shutting down.
//	select {
//	case m.queueNotification <- n:
//	case <-m.quit:
//	}
//}

//type notificationFinalityConflict struct {
//	violatingBlockHash *daghash.Hash
//}
//
//type notificationFinalityConflictResolved struct {
//	finalityBlockHash *daghash.Hash
//}
