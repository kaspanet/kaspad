package netadapter

import (
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
)

// Router routes messages by type to their respective
// input channels
type Router struct {
	inputRoutes map[string]chan<- wire.Message
	outputRoute chan<- wire.Message
}

// AddRoute registers the messages of types `messageTypes` to
// be routed to the given `inputChannel`
func (r *Router) AddRoute(messageTypes []string, inputChannel chan<- wire.Message) error {
	for _, messageType := range messageTypes {
		if _, ok := r.inputRoutes[messageType]; ok {
			return errors.Errorf("a route for '%s' already exists", messageType)
		}
		r.inputRoutes[messageType] = inputChannel
	}
	return nil
}

// RemoveRoute unregisters the messages of types `messageTypes` from
// the router
func (r *Router) RemoveRoute(messageTypes []string) error {
	for _, messageType := range messageTypes {
		if _, ok := r.inputRoutes[messageType]; !ok {
			return errors.Errorf("a route for '%s' does not exist", messageType)
		}
		delete(r.inputRoutes, messageType)
	}
	return nil
}

// RouteMessage sends the given message to the correct input
// channel as registered with AddRoute
func (r *Router) RouteMessage(message wire.Message) {
	routeInChannel := r.inputRoutes[message.Command()]
	routeInChannel <- message
}

// Close shuts down the router by closing all registered
// input channels
func (r *Router) Close() error {
	inChannels := make(map[chan<- wire.Message]struct{})
	for _, inChannel := range r.inputRoutes {
		inChannels[inChannel] = struct{}{}
	}
	for inChannel := range inChannels {
		close(inChannel)
	}
	return nil
}
