package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/kaspanet/go-secp256k1"
	"io/ioutil"
	"os"

	"github.com/kaspanet/kaspad/util"

	"github.com/pkg/errors"
)

type addressesList struct {
	myPrivateKey   *secp256k1.SchnorrKeyPair
	myAddress      util.Address
	spendAddresses []util.Address
}

func readFile() ([]byte, error) {
	addresessFilePath := activeConfig().AddressesFilePath
	return ioutil.ReadFile(addresessFilePath)
}

func loadAddresses() (*addressesList, error) {
	addressesData, err := readFile()
	if err != nil {
		return nil, err
	}

	addresses := &rawAddressesList{}
	err = json.Unmarshal(addressesData, addresses)
	if err != nil {
		return nil, err
	}

	return addresses.decode()
}

type rawAddressesList struct {
	MyPrivateKey   string   `json:"myPrivateKey"`
	MyAddress      string   `json:"myAddress"`
	SpendAddresses []string `json:"spendAddresses"`
}

func (r *rawAddressesList) decode() (*addressesList, error) {
	myKeyPair, myPublicKey, err := parsePrivateKey(r.MyPrivateKey)
	if err != nil {
		return nil, err
	}

	pubKeySerialized, err := myPublicKey.Serialize()
	if err != nil {
		panic(err)
	}

	prefix := activeConfig().ActiveNetParams.Prefix

	derivedAddr, err := util.NewAddressPubKeyHashFromPublicKey(pubKeySerialized[:], prefix)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to generate p2pkh address: %s", err)
		os.Exit(1)
	}

	myAddress, err := util.DecodeAddress(r.MyAddress, prefix)
	if err != nil {
		return nil, err
	}

	if derivedAddr.String() != myAddress.String() {
		fmt.Fprintf(os.Stderr, "myAddress %s is expected to be %s", myAddress, derivedAddr)
		os.Exit(1)
	}

	spendAddresses := make([]util.Address, 0, len(r.SpendAddresses))
	for _, rawSpendAddress := range r.SpendAddresses {
		spendAddress, err := util.DecodeAddress(rawSpendAddress, prefix)
		if err != nil {
			return nil, err
		}

		spendAddresses = append(spendAddresses, spendAddress)
	}

	addresses := &addressesList{
		myPrivateKey:   myKeyPair,
		myAddress:      myAddress,
		spendAddresses: spendAddresses,
	}

	return addresses, nil
}

func parsePrivateKey(privateKeyHex string) (*secp256k1.SchnorrKeyPair, *secp256k1.SchnorrPublicKey, error) {
	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Error parsing private key hex")
	}
	privateKey, err := secp256k1.DeserializePrivateKeyFromSlice(privateKeyBytes)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Error deserializing private key")
	}
	publicKey, err := privateKey.SchnorrPublicKey()
	if err != nil {
		return nil, nil, errors.Wrap(err, "Error generating public key")
	}
	return privateKey, publicKey, nil
}
