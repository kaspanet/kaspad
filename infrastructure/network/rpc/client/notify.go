package client

//// OnFinalityConflict is invoked when a finality conflict occurs.
//// It will only be invoked if a preceding call to
//// NotifyFinalityConflicts has been made to register for the
//// notification and the function is non-nil.
//OnFinalityConflict func(finalityConflict *model.FinalityConflictNtfn)
//
//// OnFinalityConflictResolved is invoked when a finality conflict
//// has been resolved. It will only be invoked if a preceding call to
//// NotifyFinalityConflicts has been made to register for the
//// notification and the function is non-nil.
//OnFinalityConflictResolved func(finalityBlockHash *daghash.Hash)

//case model.FinalityConflictNtfnMethod:
//	// Ignore the notification if the client is not interested in
//	// it.
//	if c.ntfnHandlers.OnFinalityConflict == nil {
//		return
//	}
//
//	finalityConflict, err := parseFinalityConflictNtfnParams(ntfn.Params)
//	if err != nil {
//		log.Warnf("Received invalid finality conflict notification: %s", err)
//		return
//	}
//
//	c.ntfnHandlers.OnFinalityConflict(finalityConflict)
//
//case model.FinalityConflictResolvedNtfnMethod:
//	// Ignore the notification if the client is not interested in
//	// it.
//	if c.ntfnHandlers.OnFinalityConflictResolved == nil {
//		return
//	}
//
//	finalityBlockHash, err := parseFinalityConflictResolvedNtfnParams(ntfn.Params)
//	if err != nil {
//		log.Warnf("Received invalid finality conflict notification: %s", err)
//		return
//	}
//
//	c.ntfnHandlers.OnFinalityConflictResolved(finalityBlockHash)

//func parseFinalityConflictNtfnParams(params []json.RawMessage) (*model.FinalityConflictNtfn, error) {
//	if len(params) != 1 {
//		return nil, wrongNumParams(len(params))
//	}
//
//	var finalityConflictNtfn model.FinalityConflictNtfn
//	err := json.Unmarshal(params[0], &finalityConflictNtfn)
//	if err != nil {
//		return nil, err
//	}
//
//	return &finalityConflictNtfn, nil
//}
//
//func parseFinalityConflictResolvedNtfnParams(params []json.RawMessage) (
//	finalityConflictBlockHash *daghash.Hash, err error) {
//
//	if len(params) != 1 {
//		return nil, wrongNumParams(len(params))
//	}
//
//	var finalityConflictResolvedNtfn model.FinalityConflictResolvedNtfn
//	err = json.Unmarshal(params[0], &finalityConflictResolvedNtfn)
//	if err != nil {
//		return nil, err
//	}
//
//	finalityBlockHash, err := daghash.NewHashFromStr(finalityConflictResolvedNtfn.FinalityBlockHash)
//	if err != nil {
//		return nil, err
//	}
//
//	return finalityBlockHash, nil
//}

//// FutureNotifyFinalityConflictsResult is a future promise to deliver the result of a
//// NotifyFinalityConflictsAsync RPC invocation (or an applicable error).
//type FutureNotifyFinalityConflictsResult chan *response
//
//// Receive waits for the response promised by the future and returns an error
//// if the registration was not successful.
//func (r FutureNotifyFinalityConflictsResult) Receive() error {
//	_, err := receiveFuture(r)
//	return err
//}
//
//// NotifyFinalityConflictsAsync returns an instance of a type that can be used to get the
//// result of the RPC at some future time by invoking the Receive function on
//// the returned instance.
////
//// See NotifyFinalityConflicts for the blocking version and more details.
//func (c *Client) NotifyFinalityConflictsAsync() FutureNotifyFinalityConflictsResult {
//	// Not supported in HTTP POST mode.
//	if c.config.HTTPPostMode {
//		return newFutureError(ErrWebsocketsRequired)
//	}
//
//	// Ignore the notification if the client is not interested in
//	// notifications.
//	if c.ntfnHandlers == nil {
//		return newNilFutureResult()
//	}
//
//	cmd := model.NewNotifyFinalityConflictsCmd()
//	return c.sendCmd(cmd)
//}
//
//// NotifyFinalityConflicts registers the client to receive notifications when
//// finality conflicts occur. The notifications are delivered to the notification
//// handlers associated with the client. Calling this function has no effect
//// if there are no notification handlers and will result in an error if the
//// client is configured to run in HTTP POST mode.
////
//// The notifications delivered as a result of this call will be via OnBlockAdded
//func (c *Client) NotifyFinalityConflicts() error {
//	return c.NotifyFinalityConflictsAsync().Receive()
//}
