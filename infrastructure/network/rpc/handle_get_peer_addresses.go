package rpc

import (
	"github.com/kaspanet/kaspad/infrastructure/network/addressmanager"
)

// handleGetPeerAddresses handles getPeerAddresses commands.
func handleGetPeerAddresses(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	return nil, nil
}

func convertAddressKeySliceToString(addressKeys []addressmanager.AddressKey) []string {
	strings := make([]string, len(addressKeys))
	for j, addr := range addressKeys {
		strings[j] = string(addr)
	}
	return strings
}
