package utxoindex

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// ScriptPublicKeyHexString is a script public key represented in hex
type ScriptPublicKeyHexString string

// UTXOOutpointEntryPairs is a map between UTXO outpoints to UTXO entries
type UTXOOutpointEntryPairs map[externalapi.DomainOutpoint]externalapi.UTXOEntry

// UTXOOutpoints is a set of UTXO outpoints
type UTXOOutpoints map[externalapi.DomainOutpoint]interface{}

// UTXOChanges is the set of changes made to the UTXO index after
// a successful update
type UTXOChanges struct {
	Added   map[ScriptPublicKeyHexString]UTXOOutpointEntryPairs
	Removed map[ScriptPublicKeyHexString]UTXOOutpoints
}
