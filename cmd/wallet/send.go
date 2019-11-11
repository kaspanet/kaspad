package main

import (
	"encoding/hex"

	"github.com/daglabs/btcd/btcec"
	"github.com/daglabs/btcd/util"
	"github.com/pkg/errors"
)

func parsePrivateKey(privateKeyHex string) (*btcec.PrivateKey, *btcec.PublicKey, error) {
	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Error parsing private key hex")
	}
	privateKey, publicKey := btcec.PrivKeyFromBytes(btcec.S256(), privateKeyBytes)
	return privateKey, publicKey, nil
}

func send(conf *sendConfig) error {
	privateKey, publicKey, err := parsePrivateKey(conf.PrivateKey)
	if err != nil {
		return err
	}

	address, err := util.NewAddressPubKeyHashFromPublicKey(publicKey.SerializeCompressed(), activeNetParams.Prefix)
	if err != nil {
		return err
	}

	utxos, err := getUTXOs(conf.APIAddress, address.String())
	if err != nil {
		return err
	}

	return nil
}
