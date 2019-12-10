package rpc

import (
	"github.com/kaspanet/kaspad/rpcmodel"
	"github.com/kaspanet/kaspad/util/subnetworkid"
)

// handleNotifyNewTransations implements the notifyNewTransactions command
// extension for websocket connections.
func handleNotifyNewTransactions(wsc *wsClient, icmd interface{}) (interface{}, error) {
	cmd, ok := icmd.(*rpcmodel.NotifyNewTransactionsCmd)
	if !ok {
		return nil, rpcmodel.ErrRPCInternal
	}

	isVerbose := cmd.Verbose != nil && *cmd.Verbose
	if isVerbose == false && cmd.Subnetwork != nil {
		return nil, &rpcmodel.RPCError{
			Code:    rpcmodel.ErrRPCInvalidParameter,
			Message: "Subnetwork switch is only allowed if verbose=true",
		}
	}

	var subnetworkID *subnetworkid.SubnetworkID
	if cmd.Subnetwork != nil {
		var err error
		subnetworkID, err = subnetworkid.NewFromStr(*cmd.Subnetwork)
		if err != nil {
			return nil, &rpcmodel.RPCError{
				Code:    rpcmodel.ErrRPCInvalidParameter,
				Message: "Subnetwork is malformed",
			}
		}
	}

	if isVerbose {
		nodeSubnetworkID := wsc.server.cfg.DAG.SubnetworkID()
		if nodeSubnetworkID.IsEqual(subnetworkid.SubnetworkIDNative) && subnetworkID != nil {
			return nil, &rpcmodel.RPCError{
				Code:    rpcmodel.ErrRPCInvalidParameter,
				Message: "Subnetwork switch is disabled when node is in Native subnetwork",
			}
		} else if nodeSubnetworkID != nil {
			if subnetworkID == nil {
				return nil, &rpcmodel.RPCError{
					Code:    rpcmodel.ErrRPCInvalidParameter,
					Message: "Subnetwork switch is required when node is partial",
				}
			}
			if !nodeSubnetworkID.IsEqual(subnetworkID) {
				return nil, &rpcmodel.RPCError{
					Code:    rpcmodel.ErrRPCInvalidParameter,
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
