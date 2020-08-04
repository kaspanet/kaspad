package flowcontext

import (
	"errors"
	"sync/atomic"

	"github.com/kaspanet/kaspad/netadapter/router"

	"github.com/kaspanet/kaspad/protocol/protocolerrors"
)

// HandleError handles an error from a flow,
// It sends the error to errChan if isStopping == 0 and increments isStopping
//
// If this is ErrRouteClosed and allowClosed = true - ignores the error
// If this is ProtocolError - logs the error
// Otherwise - panics
func (f *FlowContext) HandleError(err error, flowName string, isStopping *uint32, allowClosed bool, errChan chan<- error) {
	if errors.Is(err, router.ErrRouteClosed) {
		if allowClosed {
			return
		}
		f.handleProtocolError(err, isStopping, errChan)
		return
	}

	if protocolErr := &(protocolerrors.ProtocolError{}); !errors.As(err, &protocolErr) {
		panic(err)
	}

	log.Errorf("error from %s: %+v", flowName, err)
	f.handleProtocolError(err, isStopping, errChan)
}

func (f *FlowContext) handleProtocolError(err error, isStopping *uint32, errChan chan<- error) {
	if atomic.AddUint32(isStopping, 1) == 1 {
		errChan <- err
	}
}
