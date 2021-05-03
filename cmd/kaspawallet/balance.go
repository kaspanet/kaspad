package main

import (
	"fmt"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/keys"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet"
	"github.com/kaspanet/kaspad/util"
	"math"
)

const (
	externalKeychain = 0
	internalKeychain = 1
)

var keychains = []uint8{externalKeychain, internalKeychain}

func balance(conf *balanceConfig) error {
	client, err := connectToRPC(conf.NetParams(), conf.RPCServer)
	if err != nil {
		return err
	}

	keysFile, err := keys.ReadKeysFile(conf.NetParams(), conf.KeysFile)
	if err != nil {
		return err
	}

	var availableBalance, pendingBalance uint64
	const numAddress = math.MaxUint16
	addresses := make([]string, 0, numAddress)
	for index := 0; len(addresses) < numAddress; index++ {
		for cosignerIndex := 0; cosignerIndex < len(keysFile.ExtendedPublicKeys); cosignerIndex++ {
			for _, keychain := range keychains {
				path := fmt.Sprintf("m/%d/%d/%d", cosignerIndex, keychain, index)
				addr, err := libkaspawallet.Address(conf.NetParams(), keysFile.ExtendedPublicKeys, keysFile.MinimumSignatures, path, keysFile.ECDSA)
				if err != nil {
					return err
				}
				addresses = append(addresses, addr.String())
			}
		}
	}

	getUTXOsByAddressesResponse, err := client.GetUTXOsByAddresses(addresses)
	if err != nil {
		return err
	}

	virtualSelectedParentBlueScoreResponse, err := client.GetVirtualSelectedParentBlueScore()
	if err != nil {
		return err
	}
	virtualSelectedParentBlueScore := virtualSelectedParentBlueScoreResponse.BlueScore

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
