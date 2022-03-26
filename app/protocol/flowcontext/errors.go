package flowcontext

import (
	"errors"
	"strings"
	"sync/atomic"

	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"

	"github.com/kaspanet/kaspad/app/protocol/protocolerrors"
)

var (
	// ErrPingTimeout signifies that a ping operation timed out.
	ErrPingTimeout = protocolerrors.New(false, "timeout expired on ping")
)

// HandleError handles an error from a flow,
// It sends the error to errChan if isStopping == 0 and increments isStopping
//
// If this is ErrRouteClosed - forward it to errChan
// If this is ProtocolError - logs the error, and forward it to errChan
// Otherwise - panics
func (*FlowContext) HandleError(err error, flowName string, isStopping *uint32, errChan chan<- error) {
	isErrRouteClosed := errors.Is(err, router.ErrRouteClosed)
	if !isErrRouteClosed {
		if protocolErr := (protocolerrors.ProtocolError{}); !errors.As(err, &protocolErr) {
			panic(err)
		}
		if errors.Is(err, ErrPingTimeout) {
			// Avoid printing the call stack on ping timeouts, since users get panicked and this case is not interesting
			log.Errorf("error from %s: %s", flowName, err)
		} else {
			// Explain to the user that this is not a panic, but only a protocol error with a specific peer
			logFrame := strings.Repeat("=", 52)
			log.Errorf("Non-critical peer protocol error from %s, printing the full stack for debug purposes: \n%s\n%+v \n%s",
				flowName, logFrame, err, logFrame)
		}
	}

	if atomic.AddUint32(isStopping, 1) == 1 {
		errChan <- err
	}
}

// IsRecoverableError returns whether the error is recoverable
func (*FlowContext) IsRecoverableError(err error) bool {
	return err == nil || errors.Is(err, router.ErrRouteClosed) || errors.As(err, &protocolerrors.ProtocolError{})
}
