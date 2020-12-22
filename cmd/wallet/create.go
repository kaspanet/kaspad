package main

import (
	"fmt"

	"github.com/kaspanet/go-secp256k1"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"
)

func create() error {
	privateKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		return errors.Wrap(err, "Failed to generate private key")
	}

	fmt.Println("This is your private key, granting access to all wallet funds. Keep it safe. Use it only when sending Kaspa.")
	fmt.Printf("Private key (hex):\t%x\n\n", privateKey.Serialize()[:])

	fmt.Println("These are your public addresses for each network, where money is to be sent.")
	publicKey, err := privateKey.SchnorrPublicKey()
	if err != nil {
		return errors.Wrap(err, "Failed to generate public key")
	}
	publicKeySerialized, err := publicKey.SerializeCompressed()
	if err != nil {
		return errors.Wrap(err, "Failed to serialize public key")
	}

	for _, netParams := range []*dagconfig.Params{&dagconfig.MainnetParams, &dagconfig.TestnetParams, &dagconfig.DevnetParams} {
		addr, err := util.NewAddressPubKeyHashFromPublicKey(publicKeySerialized, netParams.Prefix)
		if err != nil {
			return errors.Wrap(err, "Failed to generate p2pkh address")
		}
		fmt.Printf("Address (%s):\t%s\n", netParams.Name, addr)
	}

	return nil
}
