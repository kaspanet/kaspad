package common

import (
	"time"

	peerpkg "github.com/c4ei/YunSeokYeol/app/protocol/peer"
	routerpkg "github.com/c4ei/YunSeokYeol/infrastructure/network/netadapter/router"

	"github.com/pkg/errors"
)

// DefaultTimeout is the default duration to wait for enqueuing/dequeuing
// to/from routes.
const DefaultTimeout = 120 * time.Second

// ErrPeerWithSameIDExists signifies that a peer with the same ID already exist.
var ErrPeerWithSameIDExists = errors.New("ready peer with the same ID already exists")

type flowExecuteFunc func(peer *peerpkg.Peer)

// Flow is a a data structure that is used in order to associate a p2p flow to some route in a router.
type Flow struct {
	Name        string
	ExecuteFunc flowExecuteFunc
}

// FlowInitializeFunc is a function that is used in order to initialize a flow
type FlowInitializeFunc func(route *routerpkg.Route, peer *peerpkg.Peer) error
