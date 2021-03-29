package main

import (
	"fmt"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/keys"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet"
	"github.com/kaspanet/kaspad/infrastructure/network/rpcclient"

	"github.com/kaspanet/kaspad/util"
)

func balance(conf *balanceConfig) error {
	client, err := rpcclient.NewRPCClient(conf.RPCServer)
	if err != nil {
		return err
	}

	keysFile, err := keys.ReadKeysFile(conf.KeysFile)
	if err != nil {
		return err
	}

	addr, err := libkaspawallet.MultiSigAddress(conf.NetParams(), keysFile.PublicKeys, keysFile.MinimumSignatures)
	if err != nil {
		return err
	}

	getUTXOsByAddressesResponse, err := client.GetUTXOsByAddresses([]string{addr.String()})
	if err != nil {
		return err
	}
	virtualSelectedParentBlueScoreResponse, err := client.GetVirtualSelectedParentBlueScore()
	if err != nil {
		return err
	}
	virtualSelectedParentBlueScore := virtualSelectedParentBlueScoreResponse.BlueScore

	var availableBalance, pendingBalance uint64
	for _, entry := range getUTXOsByAddressesResponse.Entries {
		if isUTXOSpendable(entry, virtualSelectedParentBlueScore, conf.ActiveNetParams.BlockCoinbaseMaturity) {
			availableBalance += entry.UTXOEntry.Amount
		} else {
			pendingBalance += entry.UTXOEntry.Amount
		}
	}

	fmt.Printf("Balance:\t\tKAS %f\n", float64(availableBalance)/util.SompiPerKaspa)
	if pendingBalance > 0 {
		fmt.Printf("Pending balance:\tKAS %f\n", float64(pendingBalance)/util.SompiPerKaspa)
	}

	return nil
}
