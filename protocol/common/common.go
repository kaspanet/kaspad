package common

import (
	"github.com/pkg/errors"
	"time"
)

// DefaultTimeout is the default duration to wait for enqueuing/dequeuing
// to/from routes.
const DefaultTimeout = 30 * time.Second

// ErrRouteClosed indicates that a route was closed while reading/writing.
var ErrRouteClosed = errors.New("route is closed")
