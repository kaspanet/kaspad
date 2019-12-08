package rpc

import "github.com/daglabs/kaspad/btcjson"

// API version constants
const (
	jsonrpcSemverString = "1.3.0"
	jsonrpcSemverMajor  = 1
	jsonrpcSemverMinor  = 3
	jsonrpcSemverPatch  = 0
)

// handleVersion implements the version command.
//
// NOTE: This is a btcsuite extension ported from github.com/decred/dcrd.
func handleVersion(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	result := map[string]btcjson.VersionResult{
		"btcdjsonrpcapi": {
			VersionString: jsonrpcSemverString,
			Major:         jsonrpcSemverMajor,
			Minor:         jsonrpcSemverMinor,
			Patch:         jsonrpcSemverPatch,
		},
	}
	return result, nil
}
