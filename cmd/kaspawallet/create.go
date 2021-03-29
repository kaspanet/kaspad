package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/keys"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet"
	"github.com/pkg/errors"
	"os"
)

func create(conf *createConfig) error {
	encryptedPrivateKeys, publicKeys, err := keys.CreateKeyPairs(conf.NetParams(), conf.NumPrivateKeys)
	if err != nil {
		return err
	}

	for i, publicKey := range publicKeys {
		fmt.Printf("Public key of private key #%d:\n%x\n\n", i+1, publicKey)
	}

	reader := bufio.NewReader(os.Stdin)
	for i := conf.NumPrivateKeys; i < conf.NumPublicKeys; i++ {
		fmt.Printf("Enter public key #%d here:\n", i+1)
		line, isPrefix, err := reader.ReadLine()
		if err != nil {
			return err
		}

		fmt.Println()

		if isPrefix {
			return errors.Errorf("Public key is too long")
		}

		publicKey, err := hex.DecodeString(string(line))
		if err != nil {
			return err
		}

		publicKeys = append(publicKeys, publicKey)
	}

	err = keys.WriteKeysFile(conf.KeysFile, encryptedPrivateKeys, publicKeys, conf.MinimumSignatures)
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

	fmt.Printf("The wallet address is:\n%s\n", addr)
	return nil
}
