package rpc

import (
	"fmt"
	"github.com/kaspanet/kaspad/kaspajson"
	"github.com/kaspanet/kaspad/logger"
)

// handleDebugLevel handles debugLevel commands.
func handleDebugLevel(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*kaspajson.DebugLevelCmd)

	// Special show command to list supported subsystems.
	if c.LevelSpec == "show" {
		return fmt.Sprintf("Supported subsystems %s",
			logger.SupportedSubsystems()), nil
	}

	err := logger.ParseAndSetDebugLevels(c.LevelSpec)
	if err != nil {
		return nil, &kaspajson.RPCError{
			Code:    kaspajson.ErrRPCInvalidParams.Code,
			Message: err.Error(),
		}
	}

	return "Done.", nil
}
