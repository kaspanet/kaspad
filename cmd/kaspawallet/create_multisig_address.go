package main

import (
	"encoding/hex"
	"fmt"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet"
)

func createMultisigAddress(conf *createMultisigAddressConfig) error {
	pubKeys := make([][]byte, len(conf.PublicKey))
	for i, pubKeyHex := range conf.PublicKey {
		var err error
		pubKeys[i], err = hex.DecodeString(pubKeyHex)
		if err != nil {
			return err
		}
	}

	addr, err := libkaspawallet.MultiSigAddress(conf.NetParams(), pubKeys, conf.MinimumSignatures)
	if err != nil {
		return err
	}

	fmt.Printf("Created multisig address: %s\n", addr)
	return nil
}
