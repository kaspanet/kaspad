package libkaspawallet

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/tyler-smith/go-bip39"
	"github.com/zoomy-network/zoomyd/cmd/zoomywallet/libkaspawallet/bip32"
	"github.com/zoomy-network/zoomyd/domain/dagconfig"
)

// CreateMnemonic creates a new bip-39 compatible mnemonic
func CreateMnemonic() (string, error) {
	const bip39BitSize = 256
	entropy, _ := bip39.NewEntropy(bip39BitSize)
	return bip39.NewMnemonic(entropy)
}

// Purpose and CoinType constants
const (
	SingleSignerPurpose = 44
	// Note: this is not entirely compatible to BIP 45 since
	// BIP 45 doesn't have a coin type in its derivation path.
	MultiSigPurpose = 45
	// TODO: Register the coin type in https://github.com/satoshilabs/slips/blob/master/slip-0044.md
	CoinType = 111111
)

func defaultPath(isMultisig bool) string {
	purpose := SingleSignerPurpose
	if isMultisig {
		purpose = MultiSigPurpose
	}

	return fmt.Sprintf("m/%d'/%d'/0'", purpose, CoinType)
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
