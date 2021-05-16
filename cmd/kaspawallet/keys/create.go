package keys

import (
	"bufio"
	"crypto/rand"
	"crypto/subtle"
	"fmt"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/pkg/errors"
	"github.com/tyler-smith/go-bip39"
	"os"
)

// CreateKeyPairs generates `numKeys` number of key pairs.
func CreateKeyPairs(params *dagconfig.Params, numKeys uint32, isMultisig bool) (encryptedPrivateKeys []*EncryptedMnemonic, extendedPublicKeys []string, err error) {
	return createKeyPairsFromFunction(params, numKeys, isMultisig, func(_ uint32) (string, error) {
		return libkaspawallet.CreateMnemonic()
	})
}

// ImportKeyPairs imports a `numKeys` of private keys and generates key pairs out of them.
func ImportKeyPairs(params *dagconfig.Params, numKeys uint32, isMultisig bool) (encryptedPrivateKeys []*EncryptedMnemonic, publicKeys []string, err error) {
	return createKeyPairsFromFunction(params, numKeys, isMultisig, func(keyIndex uint32) (string, error) {
		fmt.Printf("Enter mnemonic #%d here:\n", keyIndex+1)
		reader := bufio.NewReader(os.Stdin)
		mnemonic, err := reader.ReadBytes('\n')
		if err != nil {
			return "", err
		}

		if !bip39.IsMnemonicValid(string(mnemonic)) {
			return "", errors.Errorf("mnemonic is invalid")
		}

		return string(mnemonic), nil
	})
}

func createKeyPairsFromFunction(params *dagconfig.Params, numKeys uint32, isMultisig bool, keyPairFunction func(keyIndex uint32) (string, error)) (
	encryptedPrivateKeys []*EncryptedMnemonic, extendedPublicKeys []string, err error) {

	password := getPassword("Enter password for the key file:")
	confirmPassword := getPassword("Confirm password:")

	if subtle.ConstantTimeCompare(password, confirmPassword) != 1 {
		return nil, nil, errors.New("Passwords are not identical")
	}

	encryptedPrivateKeys = make([]*EncryptedMnemonic, 0, numKeys)
	for i := uint32(0); i < numKeys; i++ {
		mnemonic, err := keyPairFunction(i)
		if err != nil {
			return nil, nil, err
		}

		extendedPublicKey, err := libkaspawallet.MasterPublicKeyFromMnemonic(params, mnemonic, isMultisig)
		if err != nil {
			return nil, nil, err
		}

		extendedPublicKeys = append(extendedPublicKeys, extendedPublicKey)

		encryptedPrivateKey, err := encryptMnemonic(mnemonic, password)
		if err != nil {
			return nil, nil, err
		}
		encryptedPrivateKeys = append(encryptedPrivateKeys, encryptedPrivateKey)
	}

	return encryptedPrivateKeys, extendedPublicKeys, nil
}

func generateSalt() ([]byte, error) {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}

	return salt, nil
}

func encryptMnemonic(mnemonic string, password []byte) (*EncryptedMnemonic, error) {
	mnemonicBytes := []byte(mnemonic)

	salt, err := generateSalt()
	if err != nil {
		return nil, err
	}

	aead, err := getAEAD(password, salt)
	if err != nil {
		return nil, err
	}

	// Select a random nonce, and leave capacity for the ciphertext.
	nonce := make([]byte, aead.NonceSize(), aead.NonceSize()+len(mnemonicBytes)+aead.Overhead())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	// Encrypt the message and append the ciphertext to the nonce.
	cipher := aead.Seal(nonce, nonce, []byte(mnemonicBytes), nil)

	return &EncryptedMnemonic{
		cipher: cipher,
		salt:   salt,
	}, nil
}
