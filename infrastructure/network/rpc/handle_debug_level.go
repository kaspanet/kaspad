package rpc

import (
	"fmt"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/kaspanet/kaspad/infrastructure/network/rpc/model"
)

// handleDebugLevel handles debugLevel commands.
func handleDebugLevel(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*model.DebugLevelCmd)

	// Special show command to list supported subsystems.
	if c.LevelSpec == "show" {
		return fmt.Sprintf("Supported subsystems %s",
			logger.SupportedSubsystems()), nil
	}

	err := logger.ParseAndSetDebugLevels(c.LevelSpec)
	if err != nil {
		return nil, &model.RPCError{
			Code:    model.ErrRPCInvalidParams.Code,
			Message: err.Error(),
		}
	}

	return "Done.", nil
}
