package utxoindex

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// ScriptPublicKeyString is a script public key represented as a string
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
func ConvertScriptPublicKeyToString(scriptPublicKey []byte) ScriptPublicKeyString {
	return ScriptPublicKeyString(scriptPublicKey)
}

// ConvertStringToScriptPublicKey converts the given string to a scriptPublicKey byte slice
func ConvertStringToScriptPublicKey(string ScriptPublicKeyString) []byte {
	return []byte(string)
}
