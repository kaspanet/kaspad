package main

import (
	"bufio"
	"fmt"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet"
	"os"

	"github.com/kaspanet/kaspad/cmd/kaspawallet/keys"
	"github.com/pkg/errors"
)

func create(conf *createConfig) error {
	var encryptedMnemonics []*keys.EncryptedMnemonic
	var extendedPublicKeys []string
	var err error
	if !conf.Import {
		encryptedMnemonics, extendedPublicKeys, err = keys.CreateKeyPairs(conf.NumPrivateKeys, conf.NetParams())
	} else {
		encryptedMnemonics, extendedPublicKeys, err = keys.ImportKeyPairs(conf.NumPrivateKeys, conf.NetParams())
	}
	if err != nil {
		return err
	}

	for i, extendedPublicKey := range extendedPublicKeys {
		fmt.Printf("Extended public key of mnemonic #%d:\n%s\n\n", i+1, extendedPublicKey)
	}

	signerExtendedPublicKeys := make([]string, conf.NumPrivateKeys)
	reader := bufio.NewReader(os.Stdin)
	for i := conf.NumPrivateKeys; i < conf.NumPublicKeys; i++ {
		fmt.Printf("Enter public key #%d here:\n", i+1)
		extendedPublicKey, isPrefix, err := reader.ReadLine()
		if err != nil {
			return err
		}

		fmt.Println()

		if isPrefix {
			return errors.Errorf("Public key is too long")
		}

		signerExtendedPublicKeys[i] = string(extendedPublicKey)
		extendedPublicKeys = append(extendedPublicKeys, string(extendedPublicKey))
	}

	cosignerIndex, err := libkaspawallet.MinimumCosignerIndex(signerExtendedPublicKeys, extendedPublicKeys)
	if err != nil {
		return err
	}

	file := keys.File{
		EncryptedMnemonics:    encryptedMnemonics,
		ExtendedPublicKeys:    extendedPublicKeys,
		MinimumSignatures:     conf.MinimumSignatures,
		CosignerIndex:         cosignerIndex,
		LastUsedExternalIndex: 0,
		LastUsedInternalIndex: 0,
		ECDSA:                 conf.ECDSA,
	}
	file.SetPath(conf.NetParams(), conf.KeysFile)
	return file.Sync()
}
