package keys

import (
	"bufio"
	"crypto/cipher"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/chacha20poly1305"
	"os"
	"path/filepath"
	"runtime"
)

var (
	defaultAppDir   = util.AppDir("kaspawallet", false)
	defaultKeysFile = filepath.Join(defaultAppDir, "keys.json")
)

type keysFileJSON struct {
	EncryptedPrivateKeys []string `json:"encryptedPrivateKeys"`
	PublicKeys           []string `json:"publicKeys"`
	MinimumSignatures    uint32   `json:"minimumSignatures"`
}

type KeysFile struct {
	encryptedPrivateKeys [][]byte
	PublicKeys           [][]byte
	MinimumSignatures    uint32
}

func (kf *KeysFile) toJSON() *keysFileJSON {
	encryptedPrivateKeysHex := make([]string, len(kf.encryptedPrivateKeys))
	for i, encryptedPrivateKey := range kf.encryptedPrivateKeys {
		encryptedPrivateKeysHex[i] = hex.EncodeToString(encryptedPrivateKey)
	}

	publicKeysHex := make([]string, len(kf.PublicKeys))
	for i, publicKey := range kf.PublicKeys {
		publicKeysHex[i] = hex.EncodeToString(publicKey)
	}

	return &keysFileJSON{
		EncryptedPrivateKeys: encryptedPrivateKeysHex,
		PublicKeys:           publicKeysHex,
		MinimumSignatures:    kf.MinimumSignatures,
	}
}

func (kf *KeysFile) fromJSON(kfj *keysFileJSON) error {
	kf.MinimumSignatures = kfj.MinimumSignatures

	kf.encryptedPrivateKeys = make([][]byte, len(kfj.EncryptedPrivateKeys))
	for i, encryptedPrivateKey := range kfj.EncryptedPrivateKeys {
		var err error
		kf.encryptedPrivateKeys[i], err = hex.DecodeString(encryptedPrivateKey)
		if err != nil {
			return err
		}
	}

	kf.PublicKeys = make([][]byte, len(kfj.PublicKeys))
	for i, publicKey := range kfj.PublicKeys {
		var err error
		kf.PublicKeys[i], err = hex.DecodeString(publicKey)
		if err != nil {
			return err
		}
	}

	return nil
}

func (kf *KeysFile) DecryptPrivateKeys() ([][]byte, error) {
	password := getPassword("Password:")
	aead, err := getAEAD(password)
	if err != nil {
		return nil, err
	}

	privateKeys := make([][]byte, len(kf.encryptedPrivateKeys))
	for i, encryptedPrivateKey := range kf.encryptedPrivateKeys {
		var err error
		privateKeys[i], err = decryptPrivateKey(encryptedPrivateKey, aead)
		if err != nil {
			return nil, err
		}
	}

	return privateKeys, nil
}

func ReadKeysFile(path string) (*KeysFile, error) {
	if path == "" {
		path = defaultKeysFile
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

	keysFile := &KeysFile{}
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

func WriteKeysFile(path string, encryptedPrivateKeys [][]byte, publicKeys [][]byte, minimumSignatures uint32) error {
	if path == "" {
		path = defaultKeysFile
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

	keysFile := &KeysFile{
		encryptedPrivateKeys: encryptedPrivateKeys,
		PublicKeys:           publicKeys,
		MinimumSignatures:    minimumSignatures,
	}

	encoder := json.NewEncoder(file)
	err = encoder.Encode(keysFile.toJSON())
	if err != nil {
		return err
	}

	fmt.Printf("Wrote the keys into %s\n", path)
	return nil
}

func getAEAD(password string) (cipher.AEAD, error) {
	key := argon2.IDKey([]byte(password), []byte("kaspawallet"), 1, 64*1024, uint8(runtime.NumCPU()), 32)
	return chacha20poly1305.NewX(key)
}

func decryptPrivateKey(encryptedPrivateKey []byte, aead cipher.AEAD) ([]byte, error) {
	if len(encryptedPrivateKey) < aead.NonceSize() {
		return nil, errors.New("ciphertext too short")
	}

	// Split nonce and ciphertext.
	nonce, ciphertext := encryptedPrivateKey[:aead.NonceSize()], encryptedPrivateKey[aead.NonceSize():]

	// Decrypt the message and check it wasn't tampered with.
	return aead.Open(nil, nonce, ciphertext, nil)
}
