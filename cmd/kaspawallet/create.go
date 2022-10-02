package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet/bip32"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/utils"
	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/cmd/kaspawallet/keys"
)

func create(conf *createConfig) error {
	var encryptedMnemonics []*keys.EncryptedMnemonic
	var signerExtendedPublicKeys []string
	var err error
	isMultisig := conf.NumPublicKeys > 1
	if !conf.Import {
		encryptedMnemonics, signerExtendedPublicKeys, err = keys.CreateMnemonics(conf.NetParams(), conf.NumPrivateKeys, conf.Password, isMultisig)
	} else {
		encryptedMnemonics, signerExtendedPublicKeys, err = keys.ImportMnemonics(conf.NetParams(), conf.NumPrivateKeys, conf.Password, isMultisig)
	}
	if err != nil {
		return err
	}

	for i, extendedPublicKey := range signerExtendedPublicKeys {
		fmt.Printf("Extended public key of mnemonic #%d:\n%s\n\n", i+1, extendedPublicKey)
	}

	fmt.Printf("Notice the above is neither a secret key to your wallet " +
		"(use \"kaspawallet dump-unencrypted-data\" to see a secret seed phrase) " +
		"nor a wallet public address (use \"kaspawallet new-address\" to create and see one)\n\n")

	extendedPublicKeys := make([]string, conf.NumPrivateKeys, conf.NumPublicKeys)
	copy(extendedPublicKeys, signerExtendedPublicKeys)
	reader := bufio.NewReader(os.Stdin)
	for i := conf.NumPrivateKeys; i < conf.NumPublicKeys; i++ {
		fmt.Printf("Enter public key #%d here:\n", i+1)
		extendedPublicKey, err := utils.ReadLine(reader)
		if err != nil {
			return err
		}

		_, err = bip32.DeserializeExtendedKey(string(extendedPublicKey))
		if err != nil {
			return errors.Wrapf(err, "%s is invalid extended public key", string(extendedPublicKey))
		}

		fmt.Println()

		extendedPublicKeys = append(extendedPublicKeys, string(extendedPublicKey))
	}

	// For a read only wallet the cosigner index is 0
	cosignerIndex := uint32(0)
	if len(signerExtendedPublicKeys) > 0 {
		cosignerIndex, err = libkaspawallet.MinimumCosignerIndex(signerExtendedPublicKeys, extendedPublicKeys)
		if err != nil {
			return err
		}
	}

	file := keys.File{
		Version:            keys.LastVersion,
		EncryptedMnemonics: encryptedMnemonics,
		ExtendedPublicKeys: extendedPublicKeys,
		MinimumSignatures:  conf.MinimumSignatures,
		CosignerIndex:      cosignerIndex,
		ECDSA:              conf.ECDSA,
	}

	err = file.SetPath(conf.NetParams(), conf.KeysFile, conf.Yes)
	if err != nil {
		return err
	}

	err = file.TryLock()
	if err != nil {
		return err
	}

	err = file.Save()
	if err != nil {
		return err
	}

	fmt.Printf("Wrote the keys into %s\n", file.Path())
	return nil
}
