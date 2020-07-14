package router

import (
	"errors"
	"github.com/kaspanet/kaspad/wire"
	"time"
)

const (
	maxMessages = 100
)

// onCapacityReachedHandler is a function that is to be
// called when a route reaches capacity.
type onCapacityReachedHandler func()

// Route represents an incoming or outgoing Router route
type Route struct {
	channel chan wire.Message
	closed  bool

	onCapacityReachedHandler onCapacityReachedHandler
}

// NewRoute create a new Route
func NewRoute() *Route {
	return &Route{
		channel: make(chan wire.Message, maxMessages),
		closed:  false,
	}
}

// Enqueue enqueues a message to the Route
func (r *Route) Enqueue(message wire.Message) (isOpen bool) {
	if r.closed {
		return false
	}
	if len(r.channel) == maxMessages {
		r.onCapacityReachedHandler()
	}
	r.channel <- message
	return true
}

// Dequeue dequeues a message from the Route
func (r *Route) Dequeue() (message wire.Message, isOpen bool) {
	if r.closed {
		return nil, false
	}
	return <-r.channel, true
}

func (r *Route) EnqueueWithTimeout(message wire.Message, timeout time.Duration) (isOpen bool, err error) {
	if r.closed {
		return false, nil
	}
	if len(r.channel) == maxMessages {
		r.onCapacityReachedHandler()
	}
	select {
	case <-time.After(timeout):
		return false, errors.New("timeout expired")
	case r.channel <- message:
		return true, nil
	}
}

func (r *Route) DequeueWithTimeout(timeout time.Duration) (message wire.Message, isOpen bool, err error) {
	if r.closed {
		return nil, false, nil
	}
	select {
	case <-time.After(timeout):
		return nil, false, errors.New("timeout expired")
	case message := <-r.channel:
		return message, true, nil
	}
}

func (r *Route) setOnCapacityReachedHandler(onCapacityReachedHandler onCapacityReachedHandler) {
	r.onCapacityReachedHandler = onCapacityReachedHandler
}

// Close closes this route
func (r *Route) Close() error {
	r.closed = true
	close(r.channel)
	return nil
}
