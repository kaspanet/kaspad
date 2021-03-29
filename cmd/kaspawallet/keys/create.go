package keys

import (
	"crypto/cipher"
	"crypto/rand"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet"
	"github.com/pkg/errors"
)

// CreateKeyPairs generates `numKeys` number of key pairs.
func CreateKeyPairs(numKeys uint32) (encryptedPrivateKeys, publicKeys [][]byte, err error) {
	password := getPassword("Enter password for the key file:")
	confirmPassword := getPassword("Confirm password:")

	if password != confirmPassword {
		return nil, nil, errors.New("Passwords are not identical")
	}

	aead, err := getAEAD(password)
	if err != nil {
		return nil, nil, err
	}

	encryptedPrivateKeys = make([][]byte, 0, numKeys)
	for i := uint32(0); i < numKeys; i++ {
		privateKey, publicKey, err := libkaspawallet.CreateKeyPair()
		if err != nil {
			return nil, nil, err
		}

		publicKeys = append(publicKeys, publicKey)

		encryptedPrivateKey, err := encryptPrivateKey(privateKey, aead)
		if err != nil {
			return nil, nil, err
		}
		encryptedPrivateKeys = append(encryptedPrivateKeys, encryptedPrivateKey)
	}

	return encryptedPrivateKeys, publicKeys, nil
}

func encryptPrivateKey(privateKey []byte, aead cipher.AEAD) ([]byte, error) {
	// Select a random nonce, and leave capacity for the ciphertext.
	nonce := make([]byte, aead.NonceSize(), aead.NonceSize()+len(privateKey)+aead.Overhead())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	// Encrypt the message and append the ciphertext to the nonce.
	return aead.Seal(nonce, nonce, privateKey, nil), nil
}
