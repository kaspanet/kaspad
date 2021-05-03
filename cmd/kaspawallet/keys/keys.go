package keys

import (
	"bufio"
	"crypto/cipher"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/chacha20poly1305"
)

var (
	defaultAppDir = util.AppDir("kaspawallet", false)
)

func defaultKeysFile(netParams *dagconfig.Params) string {
	return filepath.Join(defaultAppDir, netParams.Name, "keys.json")
}

type encryptedPrivateKeyJSON struct {
	Cipher string `json:"cipher"`
	Salt   string `json:"salt"`
}

type keysFileJSON struct {
	EncryptedPrivateKeys []*encryptedPrivateKeyJSON `json:"encryptedMnemonics"`
	ExtendedPublicKeys   []string                   `json:"publicKeys"`
	MinimumSignatures    uint32                     `json:"minimumSignatures"`
	ECDSA                bool                       `json:"ecdsa"`
}

// EncryptedMnemonic represents an encrypted mnemonic
type EncryptedMnemonic struct {
	cipher []byte
	salt   []byte
}

// Data holds all the data related to the wallet keys
type Data struct {
	encryptedMnemonics []*EncryptedMnemonic
	ExtendedPublicKeys []string
	MinimumSignatures  uint32
	ECDSA              bool
}

func (d *Data) toJSON() *keysFileJSON {
	encryptedPrivateKeysJSON := make([]*encryptedPrivateKeyJSON, len(d.encryptedMnemonics))
	for i, encryptedPrivateKey := range d.encryptedMnemonics {
		encryptedPrivateKeysJSON[i] = &encryptedPrivateKeyJSON{
			Cipher: hex.EncodeToString(encryptedPrivateKey.cipher),
			Salt:   hex.EncodeToString(encryptedPrivateKey.salt),
		}
	}

	return &keysFileJSON{
		EncryptedPrivateKeys: encryptedPrivateKeysJSON,
		ExtendedPublicKeys:   d.ExtendedPublicKeys,
		MinimumSignatures:    d.MinimumSignatures,
		ECDSA:                d.ECDSA,
	}
}

func (d *Data) fromJSON(fileJSON *keysFileJSON) error {
	d.MinimumSignatures = fileJSON.MinimumSignatures
	d.ECDSA = fileJSON.ECDSA
	d.ExtendedPublicKeys = fileJSON.ExtendedPublicKeys

	d.encryptedMnemonics = make([]*EncryptedMnemonic, len(fileJSON.EncryptedPrivateKeys))
	for i, encryptedPrivateKeyJSON := range fileJSON.EncryptedPrivateKeys {
		cipher, err := hex.DecodeString(encryptedPrivateKeyJSON.Cipher)
		if err != nil {
			return err
		}

		salt, err := hex.DecodeString(encryptedPrivateKeyJSON.Salt)
		if err != nil {
			return err
		}

		d.encryptedMnemonics[i] = &EncryptedMnemonic{
			cipher: cipher,
			salt:   salt,
		}
	}

	return nil
}

// DecryptMnemonics asks the user to enter the password for the private keys and
// returns the decrypted private keys.
func (d *Data) DecryptMnemonics() ([]string, error) {
	password := getPassword("Password:")
	privateKeys := make([]string, len(d.encryptedMnemonics))
	for i, encryptedPrivateKey := range d.encryptedMnemonics {
		var err error
		privateKeys[i], err = decryptMnemonic(encryptedPrivateKey, password)
		if err != nil {
			return nil, err
		}
	}

	return privateKeys, nil
}

// ReadKeysFile returns the data related to the keys file
func ReadKeysFile(netParams *dagconfig.Params, path string) (*Data, error) {
	if path == "" {
		path = defaultKeysFile(netParams)
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	decodedFile := &keysFileJSON{}
	err = decoder.Decode(&decodedFile)
	if err != nil {
		return nil, err
	}

	keysFile := &Data{}
	err = keysFile.fromJSON(decodedFile)
	if err != nil {
		return nil, err
	}

	return keysFile, nil
}

func createFileDirectoryIfDoesntExist(path string) error {
	dir := filepath.Dir(path)
	exists, err := pathExists(dir)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	return os.MkdirAll(dir, 0700)
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)

	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil

	}

	return false, err
}

// WriteKeysFile writes a keys file with the given data
func WriteKeysFile(netParams *dagconfig.Params, path string, encryptedPrivateKeys []*EncryptedMnemonic,
	extendedPublicKeys []string, minimumSignatures uint32, ecdsa bool) error {

	if path == "" {
		path = defaultKeysFile(netParams)
	}

	exists, err := pathExists(path)
	if err != nil {
		return err
	}

	if exists {
		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("The file %s already exists. Are you sure you want to override it (type 'y' to approve)? ", path)
		line, _, err := reader.ReadLine()
		if err != nil {
			return err
		}

		if string(line) != "y" {
			return errors.Errorf("aborted keys file creation")
		}
	}

	err = createFileDirectoryIfDoesntExist(path)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	keysFile := &Data{
		encryptedMnemonics: encryptedPrivateKeys,
		ExtendedPublicKeys: extendedPublicKeys,
		MinimumSignatures:  minimumSignatures,
		ECDSA:              ecdsa,
	}

	encoder := json.NewEncoder(file)
	err = encoder.Encode(keysFile.toJSON())
	if err != nil {
		return err
	}

	fmt.Printf("Wrote the keys into %s\n", path)
	return nil
}

func getAEAD(password, salt []byte) (cipher.AEAD, error) {
	key := argon2.IDKey(password, salt, 1, 64*1024, uint8(runtime.NumCPU()), 32)
	return chacha20poly1305.NewX(key)
}

func decryptMnemonic(encryptedPrivateKey *EncryptedMnemonic, password []byte) (string, error) {
	aead, err := getAEAD(password, encryptedPrivateKey.salt)
	if err != nil {
		return "", err
	}

	if len(encryptedPrivateKey.cipher) < aead.NonceSize() {
		return "", errors.New("ciphertext too short")
	}

	// Split nonce and ciphertext.
	nonce, ciphertext := encryptedPrivateKey.cipher[:aead.NonceSize()], encryptedPrivateKey.cipher[aead.NonceSize():]

	// Decrypt the message and check it wasn't tampered with.
	decrypted, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}
