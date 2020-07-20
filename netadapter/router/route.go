package router

import (
	"sync"
	"time"

	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
)

const (
	defaultMaxMessages = 100
)

// ErrTimeout signifies that one of the router functions had a timeout.
var ErrTimeout = errors.New("timeout expired")

// onCapacityReachedHandler is a function that is to be
// called when a route reaches capacity.
type onCapacityReachedHandler func()

// Route represents an incoming or outgoing Router route
type Route struct {
	channel chan wire.Message
	// closed and closeLock are used to protect us from writing to a closed channel
	// reads use the channel's built-in mechanism to check if the channel is closed
	closed    bool
	closeLock sync.Mutex

	onCapacityReachedHandler onCapacityReachedHandler
}

// NewRoute create a new Route
func NewRoute() *Route {
	return newRouteWithCapacity(defaultMaxMessages)
}

func newRouteWithCapacity(capacity int) *Route {
	return &Route{
		channel: make(chan wire.Message, capacity),
		closed:  false,
	}
}

// Enqueue enqueues a message to the Route
func (r *Route) Enqueue(message wire.Message) (isOpen bool) {
	r.closeLock.Lock()
	defer r.closeLock.Unlock()

	if r.closed {
		return false
	}
	if len(r.channel) == defaultMaxMessages {
		r.onCapacityReachedHandler()
	}
	r.channel <- message
	return true
}

// Dequeue dequeues a message from the Route
func (r *Route) Dequeue() (message wire.Message, isOpen bool) {
	message, isOpen = <-r.channel
	return message, isOpen
}

// DequeueWithTimeout attempts to dequeue a message from the Route
// and returns an error if the given timeout expires first.
func (r *Route) DequeueWithTimeout(timeout time.Duration) (message wire.Message, isOpen bool, err error) {
	select {
	case <-time.After(timeout):
		return nil, false, errors.Wrapf(ErrTimeout, "got timeout after %s", timeout)
	case message, isOpen = <-r.channel:
		return message, isOpen, nil
	}
}

func (r *Route) setOnCapacityReachedHandler(onCapacityReachedHandler onCapacityReachedHandler) {
	r.onCapacityReachedHandler = onCapacityReachedHandler
}

// Close closes this route
func (r *Route) Close() error {
	r.closeLock.Lock()
	defer r.closeLock.Unlock()

	r.closed = true
	close(r.channel)
	return nil
}
