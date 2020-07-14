package ibd

import (
	"github.com/kaspanet/kaspad/netadapter/router"
	"sync"
)

var (
	isIBDRunning      bool
	isIBDRunningMutex sync.Mutex
	ibdStartChan      chan struct{}
)

func StartIBDIfRequired() {
	isIBDRunningMutex.Lock()
	defer isIBDRunningMutex.Unlock()

	if isIBDRunning {
		return
	}
	isIBDRunning = true
	ibdStartChan <- struct{}{}
}

func HandleIBD(incomingRoute *router.Route, outgoingRoute *router.Route) error {
	for range ibdStartChan {
		// We the flow inside a func so that the defer is called at its end
		func() {
			defer finishIBD()
		}()
	}
	return nil
}

func finishIBD() {
	isIBDRunningMutex.Lock()
	defer isIBDRunningMutex.Unlock()

	isIBDRunning = false
}
