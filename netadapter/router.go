package netadapter

import "github.com/kaspanet/kaspad/wire"

type Router struct {
	routes map[string]chan<- wire.Message
}

// AddRoute registers the messages of types `messageTypes` to
// be routed to the given `inChannel`
func (r *Router) AddRoute(messageTypes []string, inChannel chan<- wire.Message) {
	for _, messageType := range messageTypes {
		r.routes[messageType] = inChannel
	}
}

func (r *Router) RouteMessage(message wire.Message) {
	routeInChannel := r.routes[message.Command()]
	routeInChannel <- message
}
