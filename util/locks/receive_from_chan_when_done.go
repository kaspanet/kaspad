package locks

// ReceiveFromChanWhenDone takes a blocking function and returns a channel that sends an empty struct when the function is done.
func ReceiveFromChanWhenDone(callback func()) <-chan struct{} {
	ch := make(chan struct{})
	spawn("ReceiveFromChanWhenDone", func() {
		callback()
		close(ch)
	})
	return ch
}
