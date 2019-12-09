package rpc

import (
	"github.com/kaspanet/kaspad/btcjson"
	"github.com/kaspanet/kaspad/util/subnetworkid"
)

// handleGetSubnetwork handles the getSubnetwork command.
func handleGetSubnetwork(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*btcjson.GetSubnetworkCmd)

	subnetworkID, err := subnetworkid.NewFromStr(c.SubnetworkID)
	if err != nil {
		return nil, rpcDecodeHexError(c.SubnetworkID)
	}

	var gasLimit *uint64
	if !subnetworkID.IsEqual(subnetworkid.SubnetworkIDNative) &&
		!subnetworkID.IsBuiltIn() {
		limit, err := s.cfg.DAG.SubnetworkStore.GasLimit(subnetworkID)
		if err != nil {
			return nil, &btcjson.RPCError{
				Code:    btcjson.ErrRPCSubnetworkNotFound,
				Message: "Subnetwork not found.",
			}
		}
		gasLimit = &limit
	}

	subnetworkReply := &btcjson.GetSubnetworkResult{
		GasLimit: gasLimit,
	}
	return subnetworkReply, nil
}
