package flowcontext

import (
	"errors"
	"sync/atomic"

	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"

	"github.com/kaspanet/kaspad/app/protocol/protocolerrors"
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

		log.Errorf("error from %s: %s", flowName, err)
	}

	if atomic.AddUint32(isStopping, 1) == 1 {
		errChan <- err
	}
}
