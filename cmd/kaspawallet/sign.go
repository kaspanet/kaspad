package main

import (
	"encoding/hex"
	"fmt"

	"github.com/kaspanet/kaspad/cmd/kaspawallet/keys"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet"
)

func signWithPrivateKey(conf *signWithPrivateKeyConfig) error {
	partiallySignedTransaction, err := hex.DecodeString(conf.Transaction)
	if err != nil {
		return err
	}

	privateKey, err := hex.DecodeString(conf.PrivateKey)
	if err != nil {
		return err
	}

	updatedPartiallySignedTransaction, err := libkaspawallet.SignWithPrivateKey(conf.NetParams(), privateKey, partiallySignedTransaction)
	if err != nil {
		return err
	}

	isFullySigned, err := libkaspawallet.IsTransactionFullySigned(updatedPartiallySignedTransaction)
	if err != nil {
		return err
	}

	if isFullySigned {
		fmt.Println("The transaction is signed and ready to broadcast")
	} else {
		fmt.Println("Successfully signed transaction")
	}

	fmt.Printf("Transaction: %x\n", updatedPartiallySignedTransaction)
	return nil
}

func sign(conf *signConfig) error {
	keysFile, err := keys.ReadKeysFile(conf.NetParams(), conf.KeysFile)
	if err != nil {
		return err
	}

	partiallySignedTransaction, err := hex.DecodeString(conf.Transaction)
	if err != nil {
		return err
	}

	privateKeys, err := keysFile.DecryptMnemonics(conf.Password)
	if err != nil {
		return err
	}

	updatedPartiallySignedTransaction, err := libkaspawallet.Sign(conf.NetParams(), privateKeys, partiallySignedTransaction, keysFile.ECDSA)
	if err != nil {
		return err
	}

	isFullySigned, err := libkaspawallet.IsTransactionFullySigned(updatedPartiallySignedTransaction)
	if err != nil {
		return err
	}

	if isFullySigned {
		fmt.Println("The transaction is signed and ready to broadcast")
	} else {
		fmt.Println("Successfully signed transaction")
	}

	fmt.Printf("Transaction: %x\n", updatedPartiallySignedTransaction)
	return nil
}
