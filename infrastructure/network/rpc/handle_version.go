package rpc

import "github.com/kaspanet/kaspad/infrastructure/network/rpc/model"

// API version constants
const (
	jsonrpcSemverString = "1.3.0"
	jsonrpcSemverMajor  = 1
	jsonrpcSemverMinor  = 3
	jsonrpcSemverPatch  = 0
)

// handleVersion implements the version command.
func handleVersion(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	result := map[string]model.VersionResult{
		"kaspadjsonrpcapi": {
			VersionString: jsonrpcSemverString,
			Major:         jsonrpcSemverMajor,
			Minor:         jsonrpcSemverMinor,
			Patch:         jsonrpcSemverPatch,
		},
	}
	return result, nil
}
