package rpc

import (
	"github.com/kaspanet/kaspad/infrastructure/network/rpc/model"
	"github.com/kaspanet/kaspad/util/subnetworkid"
)

// handleNotifyNewTransations implements the notifyNewTransactions command
// extension for websocket connections.
func handleNotifyNewTransactions(wsc *wsClient, icmd interface{}) (interface{}, error) {
	cmd, ok := icmd.(*model.NotifyNewTransactionsCmd)
	if !ok {
		return nil, model.ErrRPCInternal
	}

	isVerbose := cmd.Verbose != nil && *cmd.Verbose
	if !isVerbose && cmd.Subnetwork != nil {
		return nil, &model.RPCError{
			Code:    model.ErrRPCInvalidParameter,
			Message: "Subnetwork switch is only allowed if verbose=true",
		}
	}

	var subnetworkID *subnetworkid.SubnetworkID
	if cmd.Subnetwork != nil {
		var err error
		subnetworkID, err = subnetworkid.NewFromStr(*cmd.Subnetwork)
		if err != nil {
			return nil, &model.RPCError{
				Code:    model.ErrRPCInvalidParameter,
				Message: "Subnetwork is malformed",
			}
		}
	}

	if isVerbose {
		nodeSubnetworkID := wsc.server.dag.SubnetworkID()
		if nodeSubnetworkID.IsEqual(subnetworkid.SubnetworkIDNative) && subnetworkID != nil {
			return nil, &model.RPCError{
				Code:    model.ErrRPCInvalidParameter,
				Message: "Subnetwork switch is disabled when node is in Native subnetwork",
			}
		} else if nodeSubnetworkID != nil {
			if subnetworkID == nil {
				return nil, &model.RPCError{
					Code:    model.ErrRPCInvalidParameter,
					Message: "Subnetwork switch is required when node is partial",
				}
			}
			if !nodeSubnetworkID.IsEqual(subnetworkID) {
				return nil, &model.RPCError{
					Code:    model.ErrRPCInvalidParameter,
					Message: "Subnetwork must equal the node's subnetwork when the node is partial",
				}
			}
		}
	}

	wsc.verboseTxUpdates = isVerbose
	wsc.subnetworkIDForTxUpdates = subnetworkID
	wsc.server.ntfnMgr.RegisterNewMempoolTxsUpdates(wsc)
	return nil, nil
}
