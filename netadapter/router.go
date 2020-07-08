package netadapter

import (
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
)

// Router routes messages by type to their respective
// input channels
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

// RouteMessage sends the given message to the correct input
// channel as registered with AddRoute
func (r *Router) RouteMessage(message wire.Message) {
	routeInChannel := r.routes[message.Command()]
	routeInChannel <- message
}
