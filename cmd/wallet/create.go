package main

import (
	"fmt"

	"github.com/kaspanet/go-secp256k1"
	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"
)

func create(conf *createConfig) error {
	privateKey, err := secp256k1.GenerateSchnorrKeyPair()
	if err != nil {
		return errors.Wrap(err, "Failed to generate private key")
	}

	fmt.Println("This is your private key, granting access to all wallet funds. Keep it safe. Use it only when sending Kaspa.")
	fmt.Printf("Private key (hex):\t%s\n\n", privateKey.SerializePrivateKey())

	fmt.Println("This is your public address, where money is to be sent.")
	publicKey, err := privateKey.SchnorrPublicKey()
	if err != nil {
		return errors.Wrap(err, "Failed to generate public key")
	}
	publicKeySerialized, err := publicKey.Serialize()
	if err != nil {
		return errors.Wrap(err, "Failed to serialize public key")
	}

	addr, err := util.NewAddressPubKeyHashFromPublicKey(publicKeySerialized[:], conf.ActiveNetParams.Prefix)
	if err != nil {
		return errors.Wrap(err, "Failed to generate p2pkh address")
	}
	fmt.Printf("Address (%s):\t%s\n", conf.ActiveNetParams.Name, addr)

	return nil
}
