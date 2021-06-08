package server

import (
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/infrastructure/network/rpcclient"
)

func connectToRPC(params *dagconfig.Params, rpcServer string) (*rpcclient.RPCClient, error) {
	rpcAddress, err := params.NormalizeRPCServerAddress(rpcServer)
	if err != nil {
		return nil, err
	}

	return rpcclient.NewRPCClient(rpcAddress)
}
