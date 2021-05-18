package main

import (
	"fmt"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/infrastructure/network/rpcclient"
	"os"
)

func isUTXOSpendable(entry *appmessage.UTXOsByAddressesEntry, virtualDAAScore uint64, coinbaseMaturity uint64) bool {
	if !entry.UTXOEntry.IsCoinbase {
		return true
	}
	blockBlueScore := entry.UTXOEntry.BlockDAAScore
	return blockBlueScore+coinbaseMaturity < virtualDAAScore
}

func printErrorAndExit(err error) {
	fmt.Fprintf(os.Stderr, "%s\n", err)
	os.Exit(1)
}

func connectToRPC(params *dagconfig.Params, rpcServer string) (*rpcclient.RPCClient, error) {
	rpcAddress, err := params.NormalizeRPCServerAddress(rpcServer)
	if err != nil {
		return nil, err
	}

	return rpcclient.NewRPCClient(rpcAddress)
}
