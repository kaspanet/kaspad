package rpc

import (
	"github.com/kaspanet/kaspad/network/rpc/model"
	"github.com/kaspanet/kaspad/util/subnetworkid"
)

// handleGetSubnetwork handles the getSubnetwork command.
func handleGetSubnetwork(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*model.GetSubnetworkCmd)

	subnetworkID, err := subnetworkid.NewFromStr(c.SubnetworkID)
	if err != nil {
		return nil, rpcDecodeHexError(c.SubnetworkID)
	}

	var gasLimit *uint64
	if !subnetworkID.IsBuiltInOrNative() {
		limit, err := s.dag.GasLimit(subnetworkID)
		if err != nil {
			return nil, &model.RPCError{
				Code:    model.ErrRPCSubnetworkNotFound,
				Message: "Subnetwork not found.",
			}
		}
		gasLimit = &limit
	}

	subnetworkReply := &model.GetSubnetworkResult{
		GasLimit: gasLimit,
	}
	return subnetworkReply, nil
}
