package rpc

import (
	"fmt"
	"github.com/kaspanet/kaspad/jsonrpc"
	"github.com/kaspanet/kaspad/logger"
)

// handleDebugLevel handles debugLevel commands.
func handleDebugLevel(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*jsonrpc.DebugLevelCmd)

	// Special show command to list supported subsystems.
	if c.LevelSpec == "show" {
		return fmt.Sprintf("Supported subsystems %s",
			logger.SupportedSubsystems()), nil
	}

	err := logger.ParseAndSetDebugLevels(c.LevelSpec)
	if err != nil {
		return nil, &jsonrpc.RPCError{
			Code:    jsonrpc.ErrRPCInvalidParams.Code,
			Message: err.Error(),
		}
	}

	return "Done.", nil
}
