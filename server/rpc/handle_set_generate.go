package rpc

import (
	"github.com/kaspanet/kaspad/config"
	"github.com/kaspanet/kaspad/rpcmodel"
)

// handleSetGenerate implements the setGenerate command.
func handleSetGenerate(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	if config.ActiveConfig().SubnetworkID != nil {
		return nil, &rpcmodel.RPCError{
			Code:    rpcmodel.ErrRPCInvalidRequest.Code,
			Message: "`setGenerate` is not supported on partial nodes.",
		}
	}

	c := cmd.(*rpcmodel.SetGenerateCmd)

	// Disable generation regardless of the provided generate flag if the
	// maximum number of threads (goroutines for our purposes) is 0.
	// Otherwise enable or disable it depending on the provided flag.
	generate := c.Generate
	genProcLimit := -1
	if c.GenProcLimit != nil {
		genProcLimit = *c.GenProcLimit
	}
	if genProcLimit == 0 {
		generate = false
	}

	if !generate {
		s.cfg.CPUMiner.Stop()
	} else {
		// Respond with an error if there are no addresses to pay the
		// created blocks to.
		if len(config.ActiveConfig().MiningAddrs) == 0 {
			return nil, &rpcmodel.RPCError{
				Code: rpcmodel.ErrRPCInternal.Code,
				Message: "No payment addresses specified " +
					"via --miningaddr",
			}
		}

		// It's safe to call start even if it's already started.
		s.cfg.CPUMiner.SetNumWorkers(int32(genProcLimit))
		s.cfg.CPUMiner.Start()
	}
	return nil, nil
}
