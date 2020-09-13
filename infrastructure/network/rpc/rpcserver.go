package rpc

//// Callback for notifications from blockdag. It notifies clients that are
//// long polling for changes or subscribed to websockets notifications.
//func (s *Server) handleBlockDAGNotification(notification *blockdag.Notification) {
//	switch notification.Type {
//	case blockdag.NTBlockAdded:
//		data, ok := notification.Data.(*blockdag.BlockAddedNotificationData)
//		if !ok {
//			log.Warnf("Block added notification data is of wrong type.")
//			break
//		}
//		block := data.Block
//
//		virtualParentsHashes := s.dag.VirtualParentHashes()
//
//		// Allow any clients performing long polling via the
//		// getBlockTemplate RPC to be notified when the new block causes
//		// their old block template to become stale.
//		s.gbtWorkState.NotifyBlockAdded(virtualParentsHashes)
//
//		// Notify registered websocket clients of incoming block.
//		s.ntfnMgr.NotifyBlockAdded(block)
//
//	case blockdag.NTChainChanged:
//		data, ok := notification.Data.(*blockdag.ChainChangedNotificationData)
//		if !ok {
//			log.Warnf("Chain changed notification data is of wrong type.")
//			break
//		}
//
//		// If the acceptance index is off we aren't capable of serving
//		// ChainChanged notifications.
//		if s.acceptanceIndex == nil {
//			break
//		}
//
//		// Notify registered websocket clients of chain changes.
//		s.ntfnMgr.NotifyChainChanged(data.RemovedChainBlockHashes,
//			data.AddedChainBlockHashes)
//
//	case blockdag.NTFinalityConflict:
//		data, ok := notification.Data.(*blockdag.FinalityConflictNotificationData)
//		if !ok {
//			log.Warnf("Finality conflict notification data is of wrong type.")
//			break
//		}
//
//		// Notify registered websocket clients of finality conflict.
//		s.ntfnMgr.NotifyFinalityConflict(data.ViolatingBlockHash)
//
//	case blockdag.NTFinalityConflictResolved:
//		data, ok := notification.Data.(*blockdag.FinalityConflictResolvedNotificationData)
//		if !ok {
//			log.Warnf("Finality conflict notification data is of wrong type.")
//			break
//		}
//
//		// Notify registered websocket clients of finality conflict resolution.
//		s.ntfnMgr.NotifyFinalityConflictResolved(data.FinalityBlockHash)
//	}
//}
