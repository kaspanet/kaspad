package utxoindex

import (
	"encoding/binary"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// ScriptPublicKeyString is a script public key represented as a string
// We use this type rather than just a byte slice because Go maps don't
// support slices as keys. See: UTXOChanges
type ScriptPublicKeyString string

// UTXOOutpointEntryPairs is a map between UTXO outpoints to UTXO entries
type UTXOOutpointEntryPairs map[externalapi.DomainOutpoint]externalapi.UTXOEntry

// UTXOOutpoints is a set of UTXO outpoints
type UTXOOutpoints map[externalapi.DomainOutpoint]interface{}

// UTXOChanges is the set of changes made to the UTXO index after
// a successful update
type UTXOChanges struct {
	Added   map[ScriptPublicKeyString]UTXOOutpointEntryPairs
	Removed map[ScriptPublicKeyString]UTXOOutpoints
}

// ConvertScriptPublicKeyToString converts the given scriptPublicKey to a string
func ConvertScriptPublicKeyToString(scriptPublicKey *externalapi.ScriptPublicKey) ScriptPublicKeyString {
	var versionBytes = make([]byte, 2) // uint16
	binary.LittleEndian.PutUint16(versionBytes, scriptPublicKey.Version)
	versionString := ScriptPublicKeyString(versionBytes)
	scriptString := ScriptPublicKeyString(scriptPublicKey.Script)
	return versionString + scriptString

}

// ConvertStringToScriptPublicKey converts the given string to a scriptPublicKey
func ConvertStringToScriptPublicKey(string ScriptPublicKeyString) *externalapi.ScriptPublicKey {
	bytes := []byte(string)
	version := binary.LittleEndian.Uint16(bytes[:2])
	script := bytes[2:]
	return &externalapi.ScriptPublicKey{Script: script, Version: version}

}
