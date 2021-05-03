package main

import (
	"fmt"

	"github.com/kaspanet/kaspad/cmd/kaspawallet/keys"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet"
)

func showAddress(conf *showAddressConfig) error {
	keysFile, err := keys.ReadKeysFile(conf.NetParams(), conf.KeysFile)
	if err != nil {
		return err
	}

	address, err := libkaspawallet.Address(conf.NetParams(), keysFile.ExtendedPublicKeys, keysFile.MinimumSignatures, keysFile.ECDSA)
	if err != nil {
		return err
	}

	fmt.Printf("The wallet address is:\n%s\n", address)
	return nil
}
