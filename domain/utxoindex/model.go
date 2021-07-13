package utxoindex

import (
	"encoding/binary"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// KeyString is a script public key represented as a string
// We use this type rather than just a byte slice because Go maps don't
// support slices as keys. See: UTXOChanges
type KeyString string

// UTXOMap is just a UTXO map
type UTXOMap map[externalapi.DomainOutpoint]externalapi.UTXOEntry

// AddressesUTXOMap Addresses to UTXOMap map
type AddressesUTXOMap map[KeyString]UTXOMap

// UTXOChanges is the set of changes made to the UTXO index after
// a successful update
type UTXOChanges struct {
	Added, Removed AddressesUTXOMap
}

// ConvertScriptPublicKeyToString converts the given scriptPublicKey to a string
func ConvertScriptPublicKeyToString(scriptPublicKey *externalapi.ScriptPublicKey) KeyString {
	versionBytes := make([]byte, 2) // uint16
	binary.LittleEndian.PutUint16(versionBytes, scriptPublicKey.Version)
	return KeyString(versionBytes) + KeyString(scriptPublicKey.Script)
}

// ConvertStringToScriptPublicKey converts the given string to a scriptPublicKey
func ConvertStringToScriptPublicKey(string KeyString) *externalapi.ScriptPublicKey {
	bytes := []byte(string)
	return &externalapi.ScriptPublicKey{Script: bytes[2:], Version: binary.LittleEndian.Uint16(bytes[:2])}
}
