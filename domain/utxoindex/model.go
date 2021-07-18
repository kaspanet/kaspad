package utxoindex

import (
	"encoding/binary"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// ScriptPublicKeyString is a script public key represented as a string
// We use this type rather than just a byte slice because Go maps don't
// support slices as keys. See: AddressesUTXOMap
type ScriptPublicKeyString string

// UTXOMap is just a UTXO map
type UTXOMap map[externalapi.DomainOutpoint]externalapi.UTXOEntry

// AddressesUTXOMap Addresses to UTXOMap map
type AddressesUTXOMap map[ScriptPublicKeyString]UTXOMap

// UTXOChanges is the set of changes made to the UTXO index after a successful update
type UTXOChanges struct {
	Added, Removed AddressesUTXOMap
}

// ConvertScriptPublicKeyToString converts the given scriptPublicKey to a string
func ConvertScriptPublicKeyToString(scriptPublicKey *externalapi.ScriptPublicKey) ScriptPublicKeyString {
	versionBytes := make([]byte, 2) // uint16
	binary.LittleEndian.PutUint16(versionBytes, scriptPublicKey.Version)
	return ScriptPublicKeyString(versionBytes) + ScriptPublicKeyString(scriptPublicKey.Script)
}
