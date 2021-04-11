package main

import (
	"encoding/hex"
	"fmt"

	"github.com/kaspanet/kaspad/cmd/kaspawallet/keys"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet"
)

func sign(conf *signConfig) error {
	keysFile, err := keys.ReadKeysFile(conf.NetParams(), conf.KeysFile)
	if err != nil {
		return err
	}

	psTxBytes, err := hex.DecodeString(conf.Transaction)
	if err != nil {
		return err
	}

	privateKeys, err := keysFile.DecryptPrivateKeys()
	if err != nil {
		return err
	}

	updatedPSTxBytes, err := libkaspawallet.Sign(privateKeys, psTxBytes, keysFile.ECDSA)
	if err != nil {
		return err
	}

	isFullySigned, err := libkaspawallet.IsTransactionFullySigned(updatedPSTxBytes)
	if err != nil {
		return err
	}

	if isFullySigned {
		fmt.Println("The transaction is signed and ready to broadcast")
	} else {
		fmt.Println("Successfully signed transaction")
	}

	fmt.Printf("Transaction: %x\n", updatedPSTxBytes)
	return nil
}
