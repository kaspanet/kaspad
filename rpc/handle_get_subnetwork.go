package rpc

import (
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/rpcmodel"
	"github.com/kaspanet/kaspad/util/subnetworkid"
)

// handleGetSubnetwork handles the getSubnetwork command.
func handleGetSubnetwork(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*rpcmodel.GetSubnetworkCmd)

	subnetworkID, err := subnetworkid.NewFromStr(c.SubnetworkID)
	if err != nil {
		return nil, rpcDecodeHexError(c.SubnetworkID)
	}

	var gasLimit *uint64
	if !subnetworkID.IsEqual(subnetworkid.SubnetworkIDNative) &&
		!subnetworkID.IsBuiltIn() {
		limit, err := blockdag.GasLimit(subnetworkID)
		if err != nil {
			return nil, &rpcmodel.RPCError{
				Code:    rpcmodel.ErrRPCSubnetworkNotFound,
				Message: "Subnetwork not found.",
			}
		}
		gasLimit = &limit
	}

	subnetworkReply := &rpcmodel.GetSubnetworkResult{
		GasLimit: gasLimit,
	}
	return subnetworkReply, nil
}
