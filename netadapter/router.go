package netadapter

import (
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
)

// OnIDReceivedHandler is a function that is to be called
// once a new Connection sends us its ID.
type OnIDReceivedHandler func(id *ID)

// Router routes messages by type to their respective
// input channels
type Router struct {
	incomingRoutes      map[string]chan<- wire.Message
	outgoingRoute       chan wire.Message
	onIDReceivedHandler OnIDReceivedHandler
}

// NewRouter creates a new empty router
func NewRouter() *Router {
	return &Router{
		incomingRoutes: make(map[string]chan<- wire.Message),
		outgoingRoute:  make(chan wire.Message),
	}
}

// SetOnIDReceivedHandler sets the onIDReceivedHandler function for
// this router
func (r *Router) SetOnIDReceivedHandler(onIDReceivedHandler OnIDReceivedHandler) {
	r.onIDReceivedHandler = onIDReceivedHandler
}

// AddRoute registers the messages of types `messageTypes` to
// be routed to the given `inputChannel`
func (r *Router) AddRoute(messageTypes []string, inputChannel chan<- wire.Message) error {
	for _, messageType := range messageTypes {
		if _, ok := r.incomingRoutes[messageType]; ok {
			return errors.Errorf("a route for '%s' already exists", messageType)
		}
		r.incomingRoutes[messageType] = inputChannel
	}
	return nil
}

// RemoveRoute unregisters the messages of types `messageTypes` from
// the router
func (r *Router) RemoveRoute(messageTypes []string) error {
	for _, messageType := range messageTypes {
		if _, ok := r.incomingRoutes[messageType]; !ok {
			return errors.Errorf("a route for '%s' does not exist", messageType)
		}
		delete(r.incomingRoutes, messageType)
	}
	return nil
}

// RouteIncomingMessage sends the given message to the correct input
// channel as registered with AddRoute
func (r *Router) RouteIncomingMessage(message wire.Message) error {
	routeInChannel, ok := r.incomingRoutes[message.Command()]
	if !ok {
		return errors.Errorf("a route for '%s' does not exist", message.Command())
	}
	routeInChannel <- message
	return nil
}

// ReadOutgoingMessage takes the next output message from
// the output channel
func (r *Router) ReadOutgoingMessage() wire.Message {
	return <-r.outgoingRoute
}

func (r *Router) WriteOutgoingMessage(message wire.Message) {
	r.outgoingRoute <- message
}

// RegisterID registers the remote connection's ID
func (r *Router) RegisterID(id *ID) {
	r.onIDReceivedHandler(id)
}

// Close shuts down the router by closing all registered
// input channels
func (r *Router) Close() error {
	inChannels := make(map[chan<- wire.Message]struct{})
	for _, inChannel := range r.incomingRoutes {
		inChannels[inChannel] = struct{}{}
	}
	for inChannel := range inChannels {
		close(inChannel)
	}
	return nil
}
