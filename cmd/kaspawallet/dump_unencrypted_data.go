package main

import (
	"bufio"
	"fmt"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/keys"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet"
	"os"

	"github.com/pkg/errors"
)

func dumpUnencryptedData(conf *dumpUnencryptedDataConfig) error {
	err := confirmDump()
	if err != nil {
		return err
	}

	keysFile, err := keys.ReadKeysFile(conf.NetParams(), conf.KeysFile)
	if err != nil {
		return err
	}

	mnemonics, err := keysFile.DecryptMnemonics()
	if err != nil {
		return err
	}

	mnemonicPublicKeys := make(map[string]struct{})
	for i, privateKey := range mnemonics {
		fmt.Printf("Mnemonic #%d:\n%x\n\n", i+1, privateKey)
		publicKey, err := libkaspawallet.ExtendedPublicKeyFromMnemonic(conf.NetParams(), privateKey, len(keysFile.ExtendedPublicKeys) > 1)
		if err != nil {
			return err
		}

		mnemonicPublicKeys[string(publicKey)] = struct{}{}
	}

	i := 1
	for _, publicKey := range keysFile.ExtendedPublicKeys {
		if _, exists := mnemonicPublicKeys[string(publicKey)]; exists {
			continue
		}

		fmt.Printf("Public key #%d:\n%x\n\n", i, publicKey)
		i++
	}

	fmt.Printf("Minimum number of signatures: %d\n", keysFile.MinimumSignatures)
	return nil
}

func confirmDump() error {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("This operation will print your unencrypted keys on the screen. Anyone that sees this information " +
		"will be able to steal your funds. Are you sure you want to proceed (y/N)? ")
	line, isPrefix, err := reader.ReadLine()
	if err != nil {
		return err
	}

	fmt.Println()

	if isPrefix || string(line) != "y" {
		return errors.Errorf("Dump aborted by user")
	}

	return nil
}
