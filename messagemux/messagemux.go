package messagemux

import "github.com/kaspanet/kaspad/wire"

// Mux represents a p2p message multiplexer.
type Mux interface {
	AddFlow(msgTypes []string, ch chan<- wire.Message)
	Stop()
}
