package locks

func TickWhenDone(callback func()) <-chan struct{} {
	ch := make(chan struct{})
	spawn(func() {
		callback()
		close(ch)
	})
	return ch
}
