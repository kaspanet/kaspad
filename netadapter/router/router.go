package router

import (
	"github.com/kaspanet/kaspad/netadapter/id"
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
)

// OnIDReceivedHandler is a function that is to be called
// once a new Connection sends us its ID.
type OnIDReceivedHandler func(id *id.ID)

// OnRouteCapacityReachedHandler is a function that is to
// be called when one of the routes reaches capacity.
type OnRouteCapacityReachedHandler func()

// Router routes messages by type to their respective
// input channels
type Router struct {
	inputRoutes                   map[string]chan wire.Message
	outputRoute                   chan wire.Message
	onIDReceivedHandler           OnIDReceivedHandler
	onRouteCapacityReachedHandler OnRouteCapacityReachedHandler
}

// NewRouter creates a new empty router
func NewRouter() *Router {
	return &Router{
		inputRoutes: make(map[string]chan wire.Message),
		outputRoute: make(chan wire.Message),
	}
}

// SetOnIDReceivedHandler sets the onIDReceivedHandler function for
// this router
func (r *Router) SetOnIDReceivedHandler(onIDReceivedHandler OnIDReceivedHandler) {
	r.onIDReceivedHandler = onIDReceivedHandler
}

// SetOnRouteCapacityReachedHandler sets the onRouteCapacityReachedHandler
// function for this router
func (r *Router) SetOnRouteCapacityReachedHandler(onRouteCapacityReachedHandler OnRouteCapacityReachedHandler) {
	r.onRouteCapacityReachedHandler = onRouteCapacityReachedHandler
}

// AddRoute registers the messages of types `messageTypes` to
// be routed to the given `inputChannel`
func (r *Router) AddRoute(messageTypes []string, inputChannel chan wire.Message) error {
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

// RouteInputMessage sends the given message to the correct input
// channel as registered with AddRoute
func (r *Router) RouteInputMessage(message wire.Message) error {
	routeInChannel, ok := r.inputRoutes[message.Command()]
	if !ok {
		return errors.Errorf("a route for '%s' does not exist", message.Command())
	}
	routeInChannel <- message
	return nil
}

// TakeOutputMessage takes the next output message from
// the output channel
func (r *Router) TakeOutputMessage() wire.Message {
	return <-r.outputRoute
}

// RegisterID registers the remote connection's ID
func (r *Router) RegisterID(id *id.ID) {
	r.onIDReceivedHandler(id)
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
