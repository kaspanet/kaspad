package common

import (
	"time"

	"github.com/pkg/errors"
)

// DefaultTimeout is the default duration to wait for enqueuing/dequeuing
// to/from routes.
const DefaultTimeout = 120 * time.Second

// ErrPeerWithSameIDExists signifies that a peer with the same ID already exist.
var ErrPeerWithSameIDExists = errors.New("ready peer with the same ID already exists")
