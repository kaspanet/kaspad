package flowcontext

import (
	"errors"
	"sync/atomic"

	"github.com/kaspanet/kaspad/netadapter/router"

	"github.com/kaspanet/kaspad/protocol/protocolerrors"
)

// HandleError handles an error from a flow,
// It increments isStopping and sends the error to errChan if isStopping == 0
//
// If this is ErrRouteClosed - ignores the error
// If this is ProtocolError - logs the error
// Otherwise - panics
func (*FlowContext) HandleError(err error, flowName string, isStopping *uint32, errChan chan<- error) {
	if errors.Is(err, router.ErrRouteClosed) {
		return
	}

	if protocolErr := &(protocolerrors.ProtocolError{}); !errors.As(err, &protocolErr) {
		panic(err)
	}

	log.Errorf("error from %s: %+v", flowName, err)
	if atomic.AddUint32(isStopping, 1) == 1 {
		errChan <- err
	}
}
