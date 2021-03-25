package main

import (
	"encoding/hex"
	"fmt"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet"
)

func sign(conf *signConfig) error {
	privateKeyBytes, err := hex.DecodeString(conf.PrivateKey)
	if err != nil {
		return err
	}

	psTxBytes, err := hex.DecodeString(conf.Transaction)
	if err != nil {
		return err
	}

	updatedPSTxBytes, err := libkaspawallet.Sign(privateKeyBytes, psTxBytes)
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

	fmt.Println("Transaction:")
	fmt.Printf(hex.EncodeToString(updatedPSTxBytes))

	return nil
}
