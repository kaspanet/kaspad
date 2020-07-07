package netadapter

import (
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
)

type Router struct {
	routes map[string]chan<- wire.Message
}

// AddRoute registers the messages of types `messageTypes` to
// be routed to the given `inChannel`
func (r *Router) AddRoute(messageTypes []string, inChannel chan<- wire.Message) error {
	for _, messageType := range messageTypes {
		if _, ok := r.routes[messageType]; ok {
			return errors.Errorf("a route for '%s' already exists", messageType)
		}
		r.routes[messageType] = inChannel
	}
	return nil
}

func (r *Router) RouteMessage(message wire.Message) {
	routeInChannel := r.routes[message.Command()]
	routeInChannel <- message
}
