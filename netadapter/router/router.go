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
	incomingRoutes map[string]*Route
	outgoingRoute  *Route

	onIDReceivedHandler           OnIDReceivedHandler
	onRouteCapacityReachedHandler OnRouteCapacityReachedHandler
}

// NewRouter creates a new empty router
func NewRouter() *Router {
	router := Router{
		incomingRoutes: make(map[string]*Route),
		outgoingRoute:  NewRoute(),
	}
	router.outgoingRoute.setOnCapacityReachedHandler(func() {
		router.onRouteCapacityReachedHandler()
	})
	return &router
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

// AddIncomingRoute registers the messages of types `messageTypes` to
// be routed to the given `route`
func (r *Router) AddIncomingRoute(messageTypes []string) (*Route, error) {
	route := NewRoute()
	for _, messageType := range messageTypes {
		if _, ok := r.incomingRoutes[messageType]; ok {
			return nil, errors.Errorf("a route for '%s' already exists", messageType)
		}
		r.incomingRoutes[messageType] = route
	}
	route.setOnCapacityReachedHandler(func() {
		r.onRouteCapacityReachedHandler()
	})
	return route, nil
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

// EnqueueIncomingMessage enqueues the given message to the
// appropriate route
func (r *Router) EnqueueIncomingMessage(message wire.Message) error {
	route, ok := r.incomingRoutes[message.Command()]
	if !ok {
		return errors.Errorf("a route for '%s' does not exist", message.Command())
	}
	return route.Enqueue(message)
}

// OutgoingRoute returns the outgoing route
func (r *Router) OutgoingRoute() *Route {
	return r.outgoingRoute
}

// RegisterID registers the remote connection's ID
func (r *Router) RegisterID(id *id.ID) {
	r.onIDReceivedHandler(id)
}

// Close shuts down the router by closing all registered
// incoming routes and the outgoing route
func (r *Router) Close() error {
	incomingRoutes := make(map[*Route]struct{})
	for _, route := range r.incomingRoutes {
		incomingRoutes[route] = struct{}{}
	}
	for route := range incomingRoutes {
		err := route.Close()
		if err != nil {
			return err
		}
	}
	return r.outgoingRoute.Close()
}
