package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/c4ei/yunseokyeol/cmd/c4exwallet/keys"
	"github.com/c4ei/yunseokyeol/cmd/c4exwallet/libc4exwallet"
	"github.com/c4ei/yunseokyeol/cmd/c4exwallet/utils"

	"github.com/pkg/errors"
)

func dumpUnencryptedData(conf *dumpUnencryptedDataConfig) error {
	if !conf.Yes {
		err := confirmDump()
		if err != nil {
			return err
		}
	}

	keysFile, err := keys.ReadKeysFile(conf.NetParams(), conf.KeysFile)
	if err != nil {
		return err
	}

	if len(conf.Password) == 0 {
		conf.Password = keys.GetPassword("Password:")
	}
	mnemonics, err := keysFile.DecryptMnemonics(conf.Password)
	if err != nil {
		return err
	}

	mnemonicPublicKeys := make(map[string]struct{})
	for i, mnemonic := range mnemonics {
		fmt.Printf("Mnemonic #%d:\n%s\n\n", i+1, mnemonic)
		publicKey, err := libc4exwallet.MasterPublicKeyFromMnemonic(conf.NetParams(), mnemonic, len(keysFile.ExtendedPublicKeys) > 1)
		if err != nil {
			return err
		}

		mnemonicPublicKeys[publicKey] = struct{}{}
	}

	i := 1
	for _, extendedPublicKey := range keysFile.ExtendedPublicKeys {
		if _, exists := mnemonicPublicKeys[extendedPublicKey]; exists {
			continue
		}

		fmt.Printf("Extended Public key #%d:\n%s\n\n", i, extendedPublicKey)
		i++
	}

	fmt.Printf("Minimum number of signatures: %d\n", keysFile.MinimumSignatures)
	return nil
}

func confirmDump() error {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("This operation will print your unencrypted keys on the screen. Anyone that sees this information " +
		"will be able to steal your funds. Are you sure you want to proceed (y/N)? ")
	line, err := utils.ReadLine(reader)
	if err != nil {
		return err
	}

	fmt.Println()

	if string(line) != "y" {
		return errors.Errorf("Dump aborted by user")
	}

	return nil
}
