package locks

// TickWhenDone takes a blocking function and returns a channel that sends an empty struct when the function is done.
func TickWhenDone(callback func()) <-chan struct{} {
	ch := make(chan struct{})
	spawn(func() {
		callback()
		close(ch)
	})
	return ch
}
