package libkaspawallet

import (
	"fmt"

	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet/bip32"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/pkg/errors"
	"github.com/tyler-smith/go-bip39"
)

// CreateMnemonic creates a new bip-39 compatible mnemonic
func CreateMnemonic() (string, error) {
	const bip39BitSize = 256
	entropy, _ := bip39.NewEntropy(bip39BitSize)
	return bip39.NewMnemonic(entropy)
}

func defaultPath(isMultisig bool) string {
	const (
		singleSignerPurpose = 44

		// Note: this is not entirely compatible to BIP 45 since
		// BIP 45 doesn't have a coin type in its derivation path.
		multiSigPurpose = 45
	)

	purpose := singleSignerPurpose
	if isMultisig {
		purpose = multiSigPurpose
	}

	// TODO: Register the coin type in https://github.com/satoshilabs/slips/blob/master/slip-0044.md
	const coinType = 111111
	return fmt.Sprintf("m/%d'/%d'/0'", purpose, coinType)
}

// MasterPublicKeyFromMnemonic returns the master public key with the correct derivation for the given mnemonic.
func MasterPublicKeyFromMnemonic(params *dagconfig.Params, mnemonic string, isMultisig bool) (string, error) {
	path := defaultPath(isMultisig)
	extendedKey, err := extendedKeyFromMnemonicAndPath(mnemonic, path, params)
	if err != nil {
		return "", err
	}

	extendedPublicKey, err := extendedKey.Public()
	if err != nil {
		return "", err
	}

	return extendedPublicKey.String(), nil
}

func extendedKeyFromMnemonicAndPath(mnemonic string, path string, params *dagconfig.Params) (*bip32.ExtendedKey, error) {
	seed := bip39.NewSeed(mnemonic, "")
	version, err := versionFromParams(params)
	if err != nil {
		return nil, err
	}

	master, err := bip32.NewMasterWithPath(seed, version, path)
	if err != nil {
		return nil, err
	}

	return master, nil
}

func versionFromParams(params *dagconfig.Params) ([4]byte, error) {
	switch params.Name {
	case dagconfig.MainnetParams.Name:
		return bip32.KaspaMainnetPrivate, nil
	case dagconfig.TestnetParams.Name:
		return bip32.KaspaTestnetPrivate, nil
	case dagconfig.DevnetParams.Name:
		return bip32.KaspaDevnetPrivate, nil
	case dagconfig.SimnetParams.Name:
		return bip32.KaspaSimnetPrivate, nil
	}

	return [4]byte{}, errors.Errorf("unknown network %s", params.Name)
}
