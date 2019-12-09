package rpc

import (
	"github.com/kaspanet/kaspad/kaspajson"
	"github.com/kaspanet/kaspad/util/subnetworkid"
)

// handleGetSubnetwork handles the getSubnetwork command.
func handleGetSubnetwork(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*kaspajson.GetSubnetworkCmd)

	subnetworkID, err := subnetworkid.NewFromStr(c.SubnetworkID)
	if err != nil {
		return nil, rpcDecodeHexError(c.SubnetworkID)
	}

	var gasLimit *uint64
	if !subnetworkID.IsEqual(subnetworkid.SubnetworkIDNative) &&
		!subnetworkID.IsBuiltIn() {
		limit, err := s.cfg.DAG.SubnetworkStore.GasLimit(subnetworkID)
		if err != nil {
			return nil, &kaspajson.RPCError{
				Code:    kaspajson.ErrRPCSubnetworkNotFound,
				Message: "Subnetwork not found.",
			}
		}
		gasLimit = &limit
	}

	subnetworkReply := &kaspajson.GetSubnetworkResult{
		GasLimit: gasLimit,
	}
	return subnetworkReply, nil
}
