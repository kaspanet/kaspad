package main

import (
	"fmt"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/keys"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet"
	"math"
)

func showAddress(conf *showAddressConfig) error {
	client, err := connectToRPC(conf.NetParams(), conf.RPCServer)
	if err != nil {
		return err
	}

	keysFile, err := keys.ReadKeysFile(conf.NetParams(), conf.KeysFile)
	if err != nil {
		return err
	}

	const numAddress = math.MaxUint16
	addresses := make([]string, 0, numAddress)
	addressIndex := make(map[string]uint32)
	for index := uint32(0); len(addresses) < numAddress; index++ {
		path := fmt.Sprintf("m/%d/%d/%d", keysFile.CosignerIndex, externalKeychain, index)
		addr, err := libkaspawallet.Address(conf.NetParams(), keysFile.ExtendedPublicKeys, keysFile.MinimumSignatures, path, keysFile.ECDSA)
		if err != nil {
			return err
		}
		addresses = append(addresses, addr.String())
		addressIndex[addr.String()] = index
	}

	getUTXOsByAddressesResponse, err := client.GetUTXOsByAddresses(addresses)
	if err != nil {
		return err
	}

	lastUsedIndex := keysFile.LastUsedIndex
	for _, entry := range getUTXOsByAddressesResponse.Entries {
		index, ok := addressIndex[entry.Address]
		if ok && index > lastUsedIndex {
			lastUsedIndex = index
		}
	}

	path := fmt.Sprintf("m/%d/%d/%d", keysFile.CosignerIndex, externalKeychain, lastUsedIndex+1)
	addr, err := libkaspawallet.Address(conf.NetParams(), keysFile.ExtendedPublicKeys, keysFile.MinimumSignatures, path, keysFile.ECDSA)
	if err != nil {
		return err
	}
	fmt.Printf("Address:\n%s\n", addr)
	return nil
}
