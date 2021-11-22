package keys

import (
	"bufio"
	"crypto/rand"
	"crypto/subtle"
	"fmt"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/utils"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/pkg/errors"
	"github.com/tyler-smith/go-bip39"
	"os"
)

// CreateMnemonics generates `numKeys` number of mnemonics.
func CreateMnemonics(params *dagconfig.Params, numKeys uint32, cmdLinePassword string, isMultisig bool) (encryptedPrivateKeys []*EncryptedMnemonic, extendedPublicKeys []string, err error) {
	mnemonics := make([]string, numKeys)
	for i := uint32(0); i < numKeys; i++ {
		var err error
		mnemonics[i], err = libkaspawallet.CreateMnemonic()
		if err != nil {
			return nil, nil, err
		}
	}

	return encryptedMnemonicExtendedPublicKeyPairs(params, mnemonics, cmdLinePassword, isMultisig)
}

// ImportMnemonics imports a `numKeys` of mnemonics.
func ImportMnemonics(params *dagconfig.Params, numKeys uint32, cmdLinePassword string, isMultisig bool) (encryptedPrivateKeys []*EncryptedMnemonic, extendedPublicKeys []string, err error) {
	mnemonics := make([]string, numKeys)
	for i := uint32(0); i < numKeys; i++ {
		fmt.Printf("Enter mnemonic #%d here:\n", i+1)
		reader := bufio.NewReader(os.Stdin)
		mnemonic, err := utils.ReadLine(reader)
		if err != nil {
			return nil, nil, err
		}

		if !bip39.IsMnemonicValid(string(mnemonic)) {
			return nil, nil, errors.Errorf("mnemonic is invalid")
		}

		mnemonics[i] = string(mnemonic)
	}
	return encryptedMnemonicExtendedPublicKeyPairs(params, mnemonics, cmdLinePassword, isMultisig)
}

func encryptedMnemonicExtendedPublicKeyPairs(params *dagconfig.Params, mnemonics []string, cmdLinePassword string, isMultisig bool) (
	encryptedPrivateKeys []*EncryptedMnemonic, extendedPublicKeys []string, err error) {
	password := []byte(cmdLinePassword)
	if len(password) == 0 {

		password = getPassword("Enter password for the key file:")
		confirmPassword := getPassword("Confirm password:")

		if subtle.ConstantTimeCompare(password, confirmPassword) != 1 {
			return nil, nil, errors.New("Passwords are not identical")
		}
	}

	encryptedPrivateKeys = make([]*EncryptedMnemonic, 0, len(mnemonics))
	for _, mnemonic := range mnemonics {
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

	aead, err := getAEAD(defaultNumThreads, password, salt)
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
