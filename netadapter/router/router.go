package router

import (
	"sync"

	"github.com/kaspanet/kaspad/domainmessage"
	"github.com/pkg/errors"
)

const outgoingRouteMaxMessages = domainmessage.MaxInvPerMsg + DefaultMaxMessages

// OnRouteCapacityReachedHandler is a function that is to
// be called when one of the routes reaches capacity.
type OnRouteCapacityReachedHandler func()

// Router routes messages by type to their respective
// input channels
type Router struct {
	incomingRoutes     map[domainmessage.MessageCommand]*Route
	incomingRoutesLock sync.RWMutex

	outgoingRoute *Route

	onRouteCapacityReachedHandler OnRouteCapacityReachedHandler
}

// NewRouter creates a new empty router
func NewRouter() *Router {
	router := Router{
		incomingRoutes: make(map[domainmessage.MessageCommand]*Route),
		outgoingRoute:  newRouteWithCapacity(outgoingRouteMaxMessages),
	}
	router.outgoingRoute.setOnCapacityReachedHandler(func() {
		router.onRouteCapacityReachedHandler()
	})
	return &router
}

// SetOnRouteCapacityReachedHandler sets the onRouteCapacityReachedHandler
// function for this router
func (r *Router) SetOnRouteCapacityReachedHandler(onRouteCapacityReachedHandler OnRouteCapacityReachedHandler) {
	r.onRouteCapacityReachedHandler = onRouteCapacityReachedHandler
}

// AddIncomingRoute registers the messages of types `messageTypes` to
// be routed to the given `route`
func (r *Router) AddIncomingRoute(messageTypes []domainmessage.MessageCommand) (*Route, error) {
	route := NewRoute()
	for _, messageType := range messageTypes {
		if r.doesIncomingRouteExist(messageType) {
			return nil, errors.Errorf("a route for '%s' already exists", messageType)
		}
		r.setIncomingRoute(messageType, route)
	}
	route.setOnCapacityReachedHandler(func() {
		r.onRouteCapacityReachedHandler()
	})
	return route, nil
}

// RemoveRoute unregisters the messages of types `messageTypes` from
// the router
func (r *Router) RemoveRoute(messageTypes []domainmessage.MessageCommand) error {
	for _, messageType := range messageTypes {
		if !r.doesIncomingRouteExist(messageType) {
			return errors.Errorf("a route for '%s' does not exist", messageType)
		}
		r.deleteIncomingRoute(messageType)
	}
	return nil
}

// EnqueueIncomingMessage enqueues the given message to the
// appropriate route
func (r *Router) EnqueueIncomingMessage(message domainmessage.Message) error {
	route, ok := r.incomingRoute(message.Command())
	if !ok {
		return errors.Errorf("a route for '%s' does not exist", message.Command())
	}
	return route.Enqueue(message)
}

// OutgoingRoute returns the outgoing route
func (r *Router) OutgoingRoute() *Route {
	return r.outgoingRoute
}

// Close shuts down the router by closing all registered
// incoming routes and the outgoing route
func (r *Router) Close() {
	r.incomingRoutesLock.Lock()
	defer r.incomingRoutesLock.Unlock()

	incomingRoutes := make(map[*Route]struct{})
	for _, route := range r.incomingRoutes {
		incomingRoutes[route] = struct{}{}
	}
	for route := range incomingRoutes {
		route.Close()
	}
	r.outgoingRoute.Close()
}

func (r *Router) incomingRoute(messageType domainmessage.MessageCommand) (*Route, bool) {
	r.incomingRoutesLock.RLock()
	defer r.incomingRoutesLock.RUnlock()

	route, ok := r.incomingRoutes[messageType]
	return route, ok
}

func (r *Router) doesIncomingRouteExist(messageType domainmessage.MessageCommand) bool {
	r.incomingRoutesLock.RLock()
	defer r.incomingRoutesLock.RUnlock()

	_, ok := r.incomingRoutes[messageType]
	return ok
}

func (r *Router) setIncomingRoute(messageType domainmessage.MessageCommand, route *Route) {
	r.incomingRoutesLock.Lock()
	defer r.incomingRoutesLock.Unlock()

	r.incomingRoutes[messageType] = route
}

func (r *Router) deleteIncomingRoute(messageType domainmessage.MessageCommand) {
	r.incomingRoutesLock.Lock()
	defer r.incomingRoutesLock.Unlock()

	delete(r.incomingRoutes, messageType)
}
