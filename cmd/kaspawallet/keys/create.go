package keys

import (
	"bufio"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet"
	"github.com/pkg/errors"
	"os"
)

// CreateKeyPairs generates `numKeys` number of key pairs.
func CreateKeyPairs(numKeys uint32, ecdsa bool) (encryptedPrivateKeys []*EncryptedPrivateKey, publicKeys [][]byte, err error) {
	return createKeyPairsFromFunction(numKeys, func(_ uint32) ([]byte, []byte, error) {
		return libkaspawallet.CreateKeyPair(ecdsa)
	})
}

// ImportKeyPairs imports a `numKeys` of private keys and generates key pairs out of them.
func ImportKeyPairs(numKeys uint32) (encryptedPrivateKeys []*EncryptedPrivateKey, publicKeys [][]byte, err error) {
	return createKeyPairsFromFunction(numKeys, func(keyIndex uint32) ([]byte, []byte, error) {
		fmt.Printf("Enter private key #%d here:\n", keyIndex+1)
		reader := bufio.NewReader(os.Stdin)
		line, isPrefix, err := reader.ReadLine()
		if err != nil {
			return nil, nil, err
		}
		if isPrefix {
			return nil, nil, errors.Errorf("Private key is too long")
		}
		privateKey, err := hex.DecodeString(string(line))
		if err != nil {
			return nil, nil, err
		}

		publicKey, err := libkaspawallet.PublicKeyFromPrivateKey(privateKey)
		if err != nil {
			return nil, nil, err
		}

		return privateKey, publicKey, nil
	})
}

func createKeyPairsFromFunction(numKeys uint32, keyPairFunction func(keyIndex uint32) ([]byte, []byte, error)) (
	encryptedPrivateKeys []*EncryptedPrivateKey, publicKeys [][]byte, err error) {

	password := getPassword("Enter password for the key file:")
	confirmPassword := getPassword("Confirm password:")

	if subtle.ConstantTimeCompare(password, confirmPassword) != 1 {
		return nil, nil, errors.New("Passwords are not identical")
	}

	encryptedPrivateKeys = make([]*EncryptedPrivateKey, 0, numKeys)
	for i := uint32(0); i < numKeys; i++ {
		privateKey, publicKey, err := keyPairFunction(i)
		if err != nil {
			return nil, nil, err
		}

		publicKeys = append(publicKeys, publicKey)

		encryptedPrivateKey, err := encryptPrivateKey(privateKey, password)
		if err != nil {
			return nil, nil, err
		}
		encryptedPrivateKeys = append(encryptedPrivateKeys, encryptedPrivateKey)
	}

	return encryptedPrivateKeys, publicKeys, nil
}

func generateSalt() ([]byte, error) {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}

	return salt, nil
}

func encryptPrivateKey(privateKey []byte, password []byte) (*EncryptedPrivateKey, error) {
	salt, err := generateSalt()
	if err != nil {
		return nil, err
	}

	aead, err := getAEAD(password, salt)
	if err != nil {
		return nil, err
	}

	// Select a random nonce, and leave capacity for the ciphertext.
	nonce := make([]byte, aead.NonceSize(), aead.NonceSize()+len(privateKey)+aead.Overhead())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	// Encrypt the message and append the ciphertext to the nonce.
	cipher := aead.Seal(nonce, nonce, privateKey, nil)

	return &EncryptedPrivateKey{
		cipher: cipher,
		salt:   salt,
	}, nil
}
